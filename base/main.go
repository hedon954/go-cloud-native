package main

import (
	"fmt"

	"net"
	"net/http"
	"strconv"

	"git.imooc.com/hedonwang/commom"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/asim/go-micro/plugins/registry/consul/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/registry"
	"github.com/jinzhu/gorm"
	"github.com/opentracing/opentracing-go"

	hystrix2 "base/plugin/hystrix"
	ratelimit "github.com/asim/go-micro/plugins/wrapper/ratelimiter/uber/v3"
	opentracing2 "github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	hostIp = "localhost" // 主机地址

	serviceName = "base" // 服务名称
	serviceHost = hostIp // 服务地址
	servicePort = "8081" // 服务端口

	// 注册中心配置
	consulHost       = hostIp
	consulPort int64 = 8500

	// 链路追踪
	tracerHost = hostIp
	tracerPort = 6831

	// 熔断降级
	hystrixHost = "0.0.0.0"
	hystrixPort = 9092

	// 服务监控
	prometheusHost = hostIp
	prometheusPort = 9192
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

	// 添加链路追踪
	tracer, closer, err := commom.NewTracer(serviceName, fmt.Sprintf("%s:%d", tracerHost, tracerPort))
	if err != nil {
		panic(err)
	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	// 添加熔断器
	hystrixStreamHandler := hystrix.NewStreamHandler()
	hystrixStreamHandler.Start()
	defer hystrixStreamHandler.Stop()

	// 启动熔断监听程序
	go func() {
		// http://192.168.1.108:9092/turbine/turbine.stream
		// 看板地址：http://localhost:9002/hystrix
		err = http.ListenAndServe(net.JoinHostPort(hystrixHost, strconv.Itoa(hystrixPort)), hystrixStreamHandler)
		if err != nil {
			panic(err)
		}
	}()

	// 创建服务
	service := micro.NewService(
		micro.Name("base"),
		micro.Version("v1"),
		micro.Registry(consulCluster),                                                 // 添加注册中心
		micro.WrapHandler(opentracing2.NewHandlerWrapper(opentracing.GlobalTracer())), // 添加链路追踪 —— 服务端模式
		micro.WrapClient(opentracing2.NewClientWrapper(opentracing.GlobalTracer())),   // 添加链路追踪 —— 客户端模式
		micro.WrapClient(hystrix2.NewClientHystrixWrapper()),                          // 添加熔断降级 —— 只作为客户端的时候起作用
		micro.WrapHandler(ratelimit.NewHandlerWrapper(1000)),                          // 添加限流：服务端模式
	)

	// 初始化服务
	service.Init()

	// 启动服务
	if err = service.Run(); err != nil {
		panic(err)
	}
}
