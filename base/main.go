package main

import (
	"fmt"
	"strconv"

	"git.imooc.com/hedonwang/commom"

	"github.com/asim/go-micro/plugins/registry/consul/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/registry"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	hostIp = "localhost" // 主机地址

	serviceHost = hostIp // 服务地址
	servicePort = "8081" // 服务端口

	// 注册中心配置
	consulHost       = hostIp
	consulPort int64 = 8500

	// 链路追踪
	tracerHost = hostIp
	tracerPort = 6831

	//hystrixPort = 9092   // 熔断端口，每个服务不能重复

	prometheusPort = 9192 // 监控端口，每个服务不能重复
)

func main() {
	//需要本地启动，mysql，consul中间件服务

	// 注册中心
	consulCluster := consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{
			consulHost + ":" + strconv.FormatInt(consulPort, 10),
		}
	})

	// 配置中心
	consulConfig, err := commom.GetConsulConfig(consulHost, consulPort, "/micro/config")
	if err != nil {
		panic(err)
	}

	// 使用配置中心连接 MySQL
	mysqlConfig, err := commom.GetMySQLFromConsul(consulConfig, "mysql")
	db, err := gorm.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
			mysqlConfig.User,
			mysqlConfig.Pwd,
			mysqlConfig.Host,
			mysqlConfig.Port,
			mysqlConfig.Database,
		))
	if err != nil {
		panic(err)
	}
	// 禁止表复数形式：user 不会自动转为 users，而是保留 user
	db.SingularTable(true)

	// 创建服务
	service := micro.NewService(
		micro.Name("base"),
		micro.Version("v1"),
		micro.Registry(consulCluster),
	)

	// 初始化服务
	service.Init()

	// 启动服务
	if err = service.Run(); err != nil {
		panic(err)
	}
}
