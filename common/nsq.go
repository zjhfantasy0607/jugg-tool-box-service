package common

import (
	"fmt"

	"github.com/nsqio/go-nsq"
	"github.com/spf13/viper"
)

var Producer *nsq.Producer

func InitProducer() {
	// 从Viper中读取地址和端口
	host := viper.GetString("nsqd.host")
	port := viper.GetString("nsqd.port")
	addr := fmt.Sprintf("%s:%s", host, port)

	// 配置NSQ
	config := nsq.NewConfig()

	// 创建生产者实例
	p, err := nsq.NewProducer(addr, config)
	if err != nil {
		panic("could not create producer: " + err.Error())
	}

	Producer = p
}
