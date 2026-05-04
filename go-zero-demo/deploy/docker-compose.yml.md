# docker-compose.yml 做什么用

这个文件用于本地准备基础依赖环境。

## 当前作用

- 启动 PostgreSQL
- 启动 Redis

## 为什么学习阶段先放这个

中小型企业项目里，服务通常不是单独运行的，往往会依赖数据库、缓存、消息队列。先把依赖环境目录预留好，后续你学到哪里就补到哪里。

## 后续可以扩展

- NATS / Kafka
- Prometheus / Grafana
- Jaeger
- MinIO

