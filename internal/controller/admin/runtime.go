package admin

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	v1 "github.com/WyRainBow/ops-portal/api/admin/v1"
	"github.com/WyRainBow/ops-portal/internal/store"
)

func (c *ControllerV1) RuntimeStatus(ctx context.Context, req *v1.RuntimeStatusReq) (res *v1.RuntimeStatusRes, err error) {
	if _, err := requireAdminOrMember(ctx); err != nil {
		return nil, err
	}
	service := req.Service
	if service == "" {
		service = "resume-backend"
	}

	db, _ := store.DB(ctx)
	dbInfo := map[string]any{"ok": false}
	if db != nil {
		t0 := time.Now()
		sqlDB, derr := db.DB()
		if derr == nil {
			derr = sqlDB.PingContext(ctx)
		}
		if derr == nil {
			dbInfo["ok"] = true
			dbInfo["dialect"] = "postgresql"
			dbInfo["latency_ms"] = float64(time.Since(t0).Milliseconds())
		} else {
			dbInfo["error"] = derr.Error()
		}
	}

	root := os.Getenv("RESUME_AGENT_ROOT")
	if root == "" {
		root = "/www/wwwroot/Resume-Agent"
	}
	gitInfo := getGitInfo(ctx, root)

	pm2Info := getPM2Info(ctx, service)
	logsInfo := pm2LogPaths(service)
	sysInfo := systemSnapshot(root)

	return &v1.RuntimeStatusRes{
		ServerTimeUTC: time.Now().UTC().Format(time.RFC3339Nano),
		Database:      dbInfo,
		Git:           gitInfo,
		System:        sysInfo,
		PM2:           pm2Info,
		Service:       map[string]any{"name": service},
		Logs:          logsInfo,
	}, nil
}

func (c *ControllerV1) RuntimeLogs(ctx context.Context, req *v1.RuntimeLogsReq) (res *v1.RuntimeLogsRes, err error) {
	if _, err := requireAdminOrMember(ctx); err != nil {
		return nil, err
	}
	service := req.Service
	if service == "" {
		service = "resume-backend"
	}
	stream := req.Stream
	if stream == "" {
		stream = "error"
	}
	lines := req.Lines
	if lines <= 0 {
		lines = 200
	}
	if lines > 2000 {
		lines = 2000
	}

	paths := pm2LogPaths(service)
	path := paths["error_path"].(string)
	if stream == "out" {
		path = paths["out_path"].(string)
	}

	content := ""
	if b, err := tailFile(path, lines, 512_000); err == nil {
		content = string(b)
	}

	return &v1.RuntimeLogsRes{
		Service: service,
		Stream:  stream,
		Lines:   lines,
		Path:    path,
		Content: content,
	}, nil
}

func getGitInfo(ctx context.Context, root string) map[string]any {
	if _, err := os.Stat(filepath.Join(root, ".git")); err != nil {
		return map[string]any{"available": false}
	}
	head := runCmd(ctx, 6*time.Second, root, "git", "rev-parse", "HEAD")
	branch := runCmd(ctx, 6*time.Second, root, "git", "rev-parse", "--abbrev-ref", "HEAD")
	dirty := runCmd(ctx, 6*time.Second, root, "git", "status", "--porcelain")
	if head["exit_code"].(int) != 0 {
		return map[string]any{"available": false}
	}
	return map[string]any{
		"available": true,
		"commit":    head["stdout"],
		"branch":    branch["stdout"],
		"dirty":     dirty["stdout"] != "",
	}
}

func pm2Home() string {
	if v := os.Getenv("PM2_HOME"); v != "" {
		return v
	}
	home, _ := os.UserHomeDir()
	if home == "" {
		home = "/root"
	}
	return filepath.Join(home, ".pm2")
}

func pm2LogPaths(service string) map[string]any {
	base := filepath.Join(pm2Home(), "logs")
	outPath := filepath.Join(base, service+"-out.log")
	errPath := filepath.Join(base, service+"-error.log")
	_, outErr := os.Stat(outPath)
	_, errErr := os.Stat(errPath)
	return map[string]any{
		"pm2_home":    pm2Home(),
		"out_path":    outPath,
		"error_path":  errPath,
		"out_exists":  outErr == nil,
		"error_exists": errErr == nil,
	}
}

func getPM2Info(ctx context.Context, service string) map[string]any {
	// pm2 jlist returns array of processes.
	j := runCmd(ctx, 8*time.Second, "", "pm2", "jlist")
	if j["exit_code"].(int) != 0 {
		return map[string]any{"ok": false, "error": j["stderr"]}
	}
	var data []map[string]any
	if err := json.Unmarshal([]byte(j["stdout"].(string)), &data); err != nil {
		return map[string]any{"ok": false, "error": "pm2 jlist parse failed: " + err.Error()}
	}
	var proc map[string]any
	for _, p := range data {
		if p["name"] == service {
			proc = p
			break
		}
	}
	out := map[string]any{"ok": true, "count": len(data)}
	if proc != nil {
		out["process"] = pm2ProcessSummary(proc)
	} else {
		out["process"] = nil
	}
	ver := runCmd(ctx, 5*time.Second, "", "pm2", "-v")
	if ver["exit_code"].(int) == 0 {
		out["version"] = ver["stdout"]
	}
	return out
}

func pm2ProcessSummary(proc map[string]any) map[string]any {
	pm2Env, _ := proc["pm2_env"].(map[string]any)
	monit, _ := proc["monit"].(map[string]any)
	s := map[string]any{
		"name": proc["name"],
		"pid":  proc["pid"],
	}
	if pm2Env != nil {
		s["status"] = pm2Env["status"]
		s["restart_time"] = pm2Env["restart_time"]
		s["pm_uptime"] = pm2Env["pm_uptime"]
		s["exec_cwd"] = pm2Env["pm_cwd"]
		if env, ok := pm2Env["env"].(map[string]any); ok {
			s["node_env"] = env["NODE_ENV"]
		}
	}
	if monit != nil {
		s["memory_bytes"] = monit["memory"]
		s["cpu_percent"] = monit["cpu"]
	}
	return s
}

func systemSnapshot(root string) map[string]any {
	out := map[string]any{}
	// disk usage via Statfs is OS-specific; keep it simple by shelling out df.
	df := runCmd(context.Background(), 4*time.Second, "", "df", "-k", root)
	out["df"] = df["stdout"]
	uptime := runCmd(context.Background(), 3*time.Second, "", "uptime")
	out["uptime"] = uptime["stdout"]
	return out
}

func runCmd(ctx context.Context, timeout time.Duration, cwd string, name string, args ...string) map[string]any {
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, name, args...)
	if cwd != "" {
		cmd.Dir = cwd
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exit := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exit = ee.ExitCode()
		} else {
			exit = 1
		}
	}
	return map[string]any{
		"exit_code": exit,
		"stdout":    stringsTrim(stdout.String()),
		"stderr":    stringsTrim(stderr.String()),
	}
}

func stringsTrim(s string) string {
	return string(bytes.TrimSpace([]byte(s)))
}

func tailFile(path string, lines int, maxBytes int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	size := st.Size()
	readSize := size
	if readSize > maxBytes {
		readSize = maxBytes
	}
	if readSize < size {
		if _, err := f.Seek(size-readSize, 0); err != nil {
			return nil, err
		}
	}
	buf := make([]byte, readSize)
	n, _ := f.Read(buf)
	buf = buf[:n]

	// Keep last N lines.
	sc := bufio.NewScanner(bytes.NewReader(buf))
	all := make([][]byte, 0)
	for sc.Scan() {
		line := append([]byte(nil), sc.Bytes()...)
		all = append(all, line)
	}
	if len(all) > lines {
		all = all[len(all)-lines:]
	}
	var out bytes.Buffer
	for _, l := range all {
		out.Write(l)
		out.WriteByte('\n')
	}
	return out.Bytes(), nil
}

var _ = strconv.IntSize
