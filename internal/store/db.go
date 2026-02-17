package store

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	dbOnce sync.Once
	dbInst *gorm.DB
	dbErr  error
)

func normalizePostgresURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	raw = strings.Replace(raw, "postgresql+psycopg2://", "postgres://", 1)
	raw = strings.Replace(raw, "postgresql+psycopg://", "postgres://", 1)
	raw = strings.Replace(raw, "postgresql://", "postgres://", 1)
	// 去掉开头多余的 "//"，避免出现 "//postgres://..." 导致 DSN 解析失败
	for strings.HasPrefix(raw, "//") {
		raw = strings.TrimPrefix(raw, "//")
	}
	// 规范化为单份 "postgres://"（如 "postgres:////user" -> "postgres://user"）
	if strings.HasPrefix(raw, "postgres://") {
		rest := strings.TrimPrefix(raw, "postgres://")
		for len(rest) > 0 && rest[0] == '/' {
			rest = strings.TrimPrefix(rest, "/")
		}
		raw = "postgres://" + rest
	}
	return raw
}

func dsnFromEnvOrCfg(ctx context.Context) (string, error) {
	if v := normalizePostgresURL(os.Getenv("OPS_PORTAL_DB_DSN")); v != "" {
		return v, nil
	}
	if v := normalizePostgresURL(os.Getenv("POSTGRESQL_URL")); v != "" {
		return v, nil
	}
	if v := normalizePostgresURL(os.Getenv("DATABASE_URL")); v != "" {
		return v, nil
	}

	v, err := g.Cfg().Get(ctx, "postgres_dsn")
	if err == nil {
		if s := normalizePostgresURL(v.String()); s != "" {
			return s, nil
		}
	}
	return "", errors.New("postgres dsn is empty: set OPS_PORTAL_DB_DSN (or POSTGRESQL_URL)")
}

func DB(ctx context.Context) (*gorm.DB, error) {
	dbOnce.Do(func() {
		dsn, err := dsnFromEnvOrCfg(ctx)
		if err != nil {
			dbErr = err
			return
		}
		gormLogger := logger.Default.LogMode(logger.Silent)

		dbInst, dbErr = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormLogger,
		})
	})
	return dbInst, dbErr
}
