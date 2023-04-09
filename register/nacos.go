package register

import (
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

func CreateNacosClient() (naming_client.INamingClient, error) {
	clientConfig := *constant.NewClientConfig(
		constant.WithNamespaceId(""), //When namespace is public, fill in the blank string here.
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("/tmp/nacos/log"),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel("debug"),
	)

	serverConfigs := []constant.ServerConfig{
		*constant.NewServerConfig(
			"127.0.0.1",
			8848,
			constant.WithScheme("http"),
			constant.WithContextPath("/nacos"),
		),
	}
	namingClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return nil, err
	}
	return namingClient, nil
}

type NacosServiceConfig struct {
	Ip          string
	Port        uint64
	ServiceName string
	ClusterName string
	GroupName   string
}

func NacosServiceRegister(namingClient naming_client.INamingClient, config NacosServiceConfig) (bool, error) {
	success, err := namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          config.Ip,
		Port:        config.Port,
		ServiceName: config.ServiceName,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    map[string]string{"idc": "shanghai"},
		//ClusterName: "cluster-a", // default value is DEFAULT
		//GroupName:   "group-a",   // default value is DEFAULT_GROUP
	})
	return success, err
}

func GetInstance(namingClient naming_client.INamingClient, serviceName string) (string, uint64, error) {
	instance, err := namingClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
		//GroupName:   "group-a",             // default value is DEFAULT_GROUP
		//Clusters:    []string{"cluster-a"}, // default value is DEFAULT
	})
	if err != nil {
		return "", 0, err
	}
	return instance.Ip, instance.Port, nil
}
