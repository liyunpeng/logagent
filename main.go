package main

import (
"fmt"
"github.com/astaxie/beego/logs"
"time"
)

func main() {
	confFile := "./conf/app.cfg"
	fmt.Println("main")
	err := initConfig(confFile)
	if err != nil {
		fmt.Printf("init conf failed:%v", err)
		return
	}
	fmt.Println("init conf success")

	err = initLogs(appConf.LogFile, appConf.LogLevel)
	if err != nil {
		fmt.Printf("init log failed:%v", err)
		return
	}
	fmt.Println("init logs success")

	timeout := time.Duration(appConf.EtcdTimeOut) * time.Second
	var etcdAddrSlice []string
	etcdAddrSlice = append(etcdAddrSlice, appConf.EtcdAddr)
	err = initEtcd(etcdAddrSlice, appConf.EtcdWatchKey, timeout)
	if err != nil {
		logs.Error("init etcd Failed:%v", err)
		return
	}
	fmt.Println("init etcd success")

	err = initKafka(appConf.KafkaAddr, appConf.ThreadNum)
	if err != nil {
		logs.Error("init kafka Failed:%v", err)
		return
	}
	fmt.Println("init kafka success")

	runServer()
}
