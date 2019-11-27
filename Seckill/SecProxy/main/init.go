package main

import (
	"Seckill/SecProxy/service"
	"context"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"
	etcd_client "go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"time"
)

var (
	redisPool  *redis.Pool
	etcdClient *etcd_client.Client
)

func initRedis()(err error){
	redisPool = &redis.Pool{
		MaxIdle:   seckillConf.RedisBlackConf.RedisMaxIdle,
		MaxActive:   seckillConf.RedisBlackConf.RedisMaxActive,
		IdleTimeout: time.Duration(seckillConf.RedisBlackConf.RedisIdleTimeout)*time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", seckillConf.RedisBlackConf.RedisAddr)
		},
	}
	conn :=redisPool.Get()
	defer conn.Close()
	_, err =conn.Do("ping")
	if err !=nil{
		logs.Error("ping redis failed,err：%v",err)
	}
	return

}

func initEtcd()(err error){
	cli, err := etcd_client.New(etcd_client.Config{
		Endpoints:   []string{seckillConf.EtcdConf.EtcdAddr},
		DialTimeout: time.Duration(seckillConf.EtcdConf.Timeout)*time.Second,
	})
	if err != nil {
		logs.Error("connect etcd failed, err:", err)
		return
	}

	etcdClient = cli
	logs.Debug("etcd start")
	return
}
//log级别转换
func convertLogLevel(level string)int{
	switch(level){
	case "debug":
		return logs.LevelDebug
	case "warn":
		return logs.LevelWarn
	case "info":
		return logs.LevelInfo
	case "trace":
		return logs.LevelTrace

	}
	return logs.LevelDebug
}
func initLogger() (err error){
	config:=make(map[string]interface{})
	config["filename"]=seckillConf.LogPath
	config["level"]=convertLogLevel(seckillConf.LogLevel)
	configStr ,err:=json.Marshal(config)
	if err != nil {
		fmt.Println("marshal failed, err:%v", err)
		return
	}

	logs.SetLogger(logs.AdapterFile,string(configStr))


	return
}

func loadSecConf()(err error){
	resp,err:=etcdClient.Get(context.Background(),seckillConf.EtcdConf.EtcdSecProductKey)
	if err!=nil{
		logs.Error("get [%s] from etcd failed,err%v",seckillConf.EtcdConf.EtcdSecProductKey,err)
		return
	}
	var secProductInfo  []service.SecProductInfoConf

	for k,v :=range resp.Kvs{
		logs.Debug("key[%s]  values[%s]",k,v)
		err = json.Unmarshal(v.Value,&secProductInfo)
		if err!=nil{
			logs.Error("Unmarshal sec product info failed,err:%v",err)
			return
		}
		logs.Debug("sec info config is [%v]",secProductInfo)
	}
	updateSecProductInfo(secProductInfo)

	return
}


func initSec()(err error){
	err = initLogger()
    if err !=nil{
    	logs.Error("ping redis failed,err:%v",err)
    	return
	}

	err = initEtcd()
	if err !=nil{
		logs.Error("init etcd failed,err:%v",err)
		return
	}
	err = initRedis()
	if err !=nil{
		logs.Error("init redis failed,err:%v",err)
		return
	}

	err = loadSecConf()
	if err !=nil{
		logs.Error("laod config failed ,err:%v",err)
		return
	}
	service.InitService(seckillConf)

	logs.Info("init sec succ")

	initSecProductWatcher()
	return
}

//监听etcd的变化
func initSecProductWatcher() {
	go watchSecProductKey(seckillConf.EtcdConf.EtcdSecProductKey)
}


func watchSecProductKey(key string) {

	cli, err := etcd_client.New(etcd_client.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		logs.Error("connect etcd failed, err:", err)
		return
	}

	logs.Debug("begin watch key:%s", key)
	for {
		rch := cli.Watch(context.Background(), key)
		var secProductInfo [] service.SecProductInfoConf
		var getConfSucc = true

		for wresp := range rch {
			for _, ev := range wresp.Events {
				if ev.Type == mvccpb.DELETE {
					logs.Warn("key[%s] 's config deleted", key)
					continue
				}

				if ev.Type == mvccpb.PUT && string(ev.Kv.Key) == key {
					err = json.Unmarshal(ev.Kv.Value, &secProductInfo)
					if err != nil {
						logs.Error("key [%s], Unmarshal[%s], err:%v ", err)
						getConfSucc = false
						continue
					}
				}
				logs.Debug("get config from etcd, %s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
			}

			if getConfSucc {
				logs.Debug("get config from etcd succ, %v", secProductInfo)
				updateSecProductInfo(secProductInfo)
			}
		}

	}
}

func updateSecProductInfo(secProductInfo []service.SecProductInfoConf) {
	var tmp map[int]*service.SecProductInfoConf = make(map[int]*service.SecProductInfoConf,1024)
	for _,v:=range secProductInfo{
		ProductInfo := v
		tmp[v.ProductId]=&ProductInfo

	}
	seckillConf.RWSecProductLock.Lock()
	seckillConf.SecProductInfoMap = tmp
	seckillConf.RWSecProductLock.Unlock()
}



