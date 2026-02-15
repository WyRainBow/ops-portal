package store

import (
	"context"
	"os"
	"sync"
	"time"

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

func dsnFromEnvOrCfg(ctx context.Context) (string, error) {
	if v := os.Getenv("OPS_PORTAL_DB_DSN"); v != "" {
		return v, nil
	}
	v, err := g.Cfg().Get(ctx, "postgres_dsn")
	if err != nil {
		return "", err
	}
	return v.String(), nil
}

func DB(ctx context.Context) (*gorm.DB, error) {
	dbOnce.Do(func() {
		dsn, err := dsnFromEnvOrCfg(ctx)
		if err != nil {
			dbErr = err
			return
		}
		// Keep logging minimal by default; can be configured later.
		gormLogger := logger.New(
			g.Log(),
			logger.Config{
				SlowThreshold:             500 * time.Millisecond,
				LogLevel:                  logger.Silent,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		)

		dbInst, dbErr = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormLogger,
		})
	})
	return dbInst, dbErr
}

