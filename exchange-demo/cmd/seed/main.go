// cmd/seed 向数据库写入模拟用户数据。
//
// 用法：
//
//	go run ./cmd/seed
//
// 幂等：已存在的用户跳过，不会重复插入。
package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/anthdm/crypto-exchange/db"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// UserModel 仅在 seed 内部用，避免循环引用 server 包。
type UserModel struct {
	gorm.Model
	UserID  int64   `gorm:"uniqueIndex;not null"`
	Balance float64 `gorm:"default:0"`
}

func (UserModel) TableName() string { return "users" }

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://cex:cex_secret@localhost:5432/cex_db?sslmode=disable&TimeZone=Asia/Shanghai"
	}

	gormDB, err := db.Connect(dsn)
	if err != nil {
		logrus.WithError(err).Fatal("connect db failed")
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	created := 0
	skipped := 0

	// UserID 100-199：留给模拟交易者（避开 7/8/666 这几个系统用户）
	for i := int64(100); i < 200; i++ {
		// 随机初始余额：5000 ~ 50000
		balance := 5000.0 + rng.Float64()*45000.0

		result := gormDB.Where(UserModel{UserID: i}).FirstOrCreate(&UserModel{
			UserID:  i,
			Balance: balance,
		})

		if result.Error != nil {
			logrus.WithError(result.Error).Errorf("failed to seed user %d", i)
			continue
		}

		if result.RowsAffected == 0 {
			skipped++
		} else {
			created++
			logrus.WithFields(logrus.Fields{
				"userID":  i,
				"balance": balance,
			}).Info("seeded user")
		}
	}

	logrus.WithFields(logrus.Fields{
		"created": created,
		"skipped": skipped,
	}).Info("seed complete")
}
