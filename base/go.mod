module base

go 1.16

require (
	git.imooc.com/hedonwang/commom v0.0.0-20221018024055-6ae991ffb6c1
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/asim/go-micro/plugins/registry/consul/v3 v3.7.0
	github.com/asim/go-micro/v3 v3.7.1
	github.com/jinzhu/gorm v1.9.16
	google.golang.org/protobuf v1.27.1
	k8s.io/api v0.22.4 //其它版本会报错
	k8s.io/client-go v0.22.4 //其它版本会报错
)
