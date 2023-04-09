package register

import (
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"time"
)

// MsRegister 注册中心, 服务名称获取服务的地址
type MsRegister interface {
	CreateClient()
	ServiceRegister(serviceName string, config any)
	GetService(serviceName string)
	Close() error
}

type MsRegisterOption struct {
	Endpoints    []string      //节点
	DialTimeout  time.Duration //超时时间
	ServiceName  string
	Host         string
	Port         int
	NamespaceId  string
	TimeoutMs    uint64
	LogLevel     string
	ServerConfig []constant.ServerConfig
}
