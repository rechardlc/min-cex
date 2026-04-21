DSN ?= "postgres://cex:cex_secret@localhost:5432/cex_db?sslmode=disable"

build:
	go build -o bin/exchange.exe .

run: build
	./bin/exchange.exe

test:
	go test -v ./...

# ── 数据库 ────────────────────────────────────────────────────────────────────

# 启动 PostgreSQL（后台）
db:
	docker compose up -d postgres

# 停止并删除容器（保留数据卷）
db-down:
	docker compose down

# 停止并删除容器 + 数据卷（危险：清空所有数据）
db-reset:
	docker compose down -v

# 执行所有待执行迁移（up）
migrate-up:
	go run ./cmd/migrate -action=up

# 回滚一个版本
migrate-down:
	go run ./cmd/migrate -action=down

# 强制设置版本号（修复 dirty 状态）：make migrate-force VERSION=1
migrate-force:
	go run ./cmd/migrate -action=force -version=$(VERSION)

# 查看当前迁移版本
migrate-version:
	go run ./cmd/migrate -action=version

# 写入模拟用户数据（UserID 100-199，幂等）
seed:
	go run ./cmd/seed

# ── 开发 ──────────────────────────────────────────────────────────────────────

# 热更新开发模式（需要先 make db && make migrate-up）
dev:
	air

# 安装 air 热更新工具
install-air:
	go install github.com/air-verse/air@latest

# 清理编译产物
clean:
	rm -rf bin/ tmp/

.PHONY: build run test db db-down db-reset migrate-up migrate-down migrate-force migrate-version dev install-air clean