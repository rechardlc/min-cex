# healthhandler.go 做什么用

`handler` 层负责接住 HTTP 请求。

## 在真实项目里的职责

- 读取请求参数
- 调用 `logic`
- 返回统一响应

## 这里为什么不建议写复杂逻辑

因为 `handler` 应该尽量薄。复杂业务放在 `logic`，这样更利于测试和维护。

