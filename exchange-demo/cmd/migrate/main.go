// cmd/migrate 是数据库迁移的独立 CLI 工具。
//
// 用法：
//
//	go run ./cmd/migrate -action=up
//	go run ./cmd/migrate -action=down
//	go run ./cmd/migrate -action=force -version=2
//	go run ./cmd/migrate -action=version
//
// DSN 优先从环境变量 DATABASE_URL 读取。
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/anthdm/crypto-exchange/db"
	"github.com/sirupsen/logrus"
)

func main() {
	action := flag.String("action", "up", "migrate action: up | down | force | version")
	version := flag.Int("version", 0, "target version (only for action=force)")
	flag.Parse()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://cex:cex_secret@localhost:5432/cex_db?sslmode=disable&TimeZone=Asia/Shanghai"
	}

	var err error
	switch *action {
	case "up":
		err = db.MigrateUp(dsn)
	case "down":
		err = db.MigrateDown(dsn)
	case "force":
		if *version == 0 {
			logrus.Fatal("force action requires -version flag, e.g. -version=1")
		}
		err = db.MigrateForce(dsn, *version)
	case "version":
		v, dirty, verErr := db.MigrateVersion(dsn)
		if verErr != nil {
			logrus.WithError(verErr).Fatal("get version failed")
		}
		fmt.Printf("current version: %d, dirty: %v\n", v, dirty)
		return
	default:
		logrus.Fatalf("unknown action: %s (up|down|force|version)", *action)
	}

	if err != nil {
		logrus.WithError(err).Fatalf("migrate %s failed", *action)
	}
}
