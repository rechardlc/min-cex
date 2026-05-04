package svc

import "go-zero-demo/apps/gateway/internal/config"

type ServiceContext struct {
    Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
    return &ServiceContext{Config: c}
}

