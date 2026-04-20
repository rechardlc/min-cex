// Package db 提供数据库连接和迁移工具。
//
// 使用 embed.FS 将 migrations/ 目录下的 SQL 文件打进二进制，
// 确保部署时不需要额外携带迁移文件。
package db

import (
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Connect 打开 GORM 数据库连接（不执行迁移）。
func Connect(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	return db, nil
}

// MigrateUp 执行所有待执行的 up 迁移。
// 如果已是最新版本（ErrNoChange），不报错，正常返回。
func MigrateUp(dsn string) error {
	m, err := newMigrator(dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}

	version, dirty, _ := m.Version()
	logrus.WithFields(logrus.Fields{
		"version": version,
		"dirty":   dirty,
	}).Info("migrations applied")

	return nil
}

// MigrateDown 回滚一个版本。
func MigrateDown(dsn string) error {
	m, err := newMigrator(dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate down: %w", err)
	}

	version, dirty, _ := m.Version()
	logrus.WithFields(logrus.Fields{
		"version": version,
		"dirty":   dirty,
	}).Info("rolled back one step")

	return nil
}

// MigrateForce 强制设置版本号（用于修复 dirty 状态）。
// dirty=true 表示上次迁移执行到一半失败了，需要手动 force 修复。
func MigrateForce(dsn string, version int) error {
	m, err := newMigrator(dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Force(version); err != nil {
		return fmt.Errorf("migrate force v%d: %w", version, err)
	}

	logrus.WithField("version", version).Info("forced migration version")
	return nil
}

// MigrateVersion 打印当前迁移版本。
func MigrateVersion(dsn string) (uint, bool, error) {
	m, err := newMigrator(dsn)
	if err != nil {
		return 0, false, err
	}
	defer m.Close()
	return m.Version()
}

// newMigrator 创建 migrate 实例，SQL 来源是 embed.FS。
func newMigrator(dsn string) (*migrate.Migrate, error) {
	// iofs：把 embed.FS 适配成 migrate 的 source.Driver 接口
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("iofs source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, dsn)
	if err != nil {
		return nil, fmt.Errorf("new migrator: %w", err)
	}

	return m, nil
}
