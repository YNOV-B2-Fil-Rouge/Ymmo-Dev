// Package database wires the application to MariaDB through GORM.
// Keeping the connection logic in one place follows the Single
// Responsibility Principle: nothing else in the app knows *how* we
// connect, only that it receives a ready-to-use *gorm.DB.
package database

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"ymmo/internal/config"
)

// Connect opens a pooled connection to MariaDB and verifies it with a
// ping. It returns an error instead of panicking so the caller decides
// how to handle a failed startup.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	gormLog := logger.Default.LogMode(logger.Warn)
	if cfg.AppEnv == "development" {
		gormLog = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(mysql.Open(cfg.DB.DSN()), &gorm.Config{
		Logger: gormLog,
	})
	if err != nil {
		return nil, fmt.Errorf("open mariadb: %w", err)
	}

	// Tune the underlying connection pool for predictable behaviour.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("access sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping mariadb: %w", err)
	}

	return db, nil
}
