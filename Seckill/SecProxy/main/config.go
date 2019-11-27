package main

import (
	"Seckill/SecProxy/service"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"strings"
)

var(
	seckillConf = &service.SecSkillConf{
		SecProductInfoMap:make(map[int]*service.SecProductInfoConf,1024),
	}
)

func initConfig() (err error) {
	//---------------------------etcd config read----------------------
	etcdAddr := beego.AppConfig.String("etcd_addr")
	logs.Debug("read config succ, etcd addr:%v", etcdAddr)
	seckillConf.EtcdConf.EtcdAddr = etcdAddr
	//---------------------------------------------------------------------

	//-------------------------redis black------------------------------------------


	redisBlackAddr := beego.AppConfig.String("redis_black_addr")
	logs.Debug("read config succ, redis addr:%v", redisBlackAddr)
	seckillConf.RedisBlackConf.RedisAddr = redisBlackAddr


	redisMaxIdle,err := beego.AppConfig.Int("redis_black_max_idle")
	if err !=nil{
		err = fmt.Errorf("init config failed,read redis_black_max_idle error:%v",err)
	}
	redisMaxActive,err := beego.AppConfig.Int("redis_black_max_active")
	if err !=nil{
		err = fmt.Errorf("init config failed,read redis_black_max_active error:%v",err)
	}
	redisIdleTimeout,err := beego.AppConfig.Int("redis_black_idle_timeout")
	if err !=nil{
		err = fmt.Errorf("init config failed,read redis_black_idle_timeout error:%v",err)
	}
//-----------------------------------------------------------------------------------------

//--------------------redis proxy----------------------------------------------------------------------
	redisProxy2LayerAddr := beego.AppConfig.String("redis_proxy2layer_addr")
	logs.Debug("read config succ, redisProxy2LayerAddr:%v", redisProxy2LayerAddr)
	seckillConf.RedisProxy2LayerConf.RedisAddr = redisProxy2LayerAddr


	redisProxy2LayerMaxIdle,err := beego.AppConfig.Int("redis_proxy2layer_max_idle")
	if err !=nil{
		err = fmt.Errorf("init config failed,read redis_black_max_idle error:%v",err)
	}
	redisProxy2LayerMaxActive,err := beego.AppConfig.Int("redis_proxy2layer_max_active")
	if err !=nil{
		err = fmt.Errorf("init config failed,read redis_black_max_active error:%v",err)
	}
	redisProxy2LayerIdleTimeout,err := beego.AppConfig.Int("redis_proxy2layer_idle_timeout")
	if err !=nil{
		err = fmt.Errorf("init config failed,read redis_black_idle_timeout error:%v",err)
	}
	//-----------------------------------------------------------------------------------------

	if len(redisBlackAddr) == 0 || len(etcdAddr) == 0 {
		err = fmt.Errorf("init config failed, redis[%s] or etcd[%s] config is null", redisBlackAddr, etcdAddr)
		return
	}
	seckillConf.RedisBlackConf.RedisMaxIdle=redisMaxIdle
	seckillConf.RedisBlackConf.RedisMaxActive=redisMaxActive
	seckillConf.RedisBlackConf.RedisIdleTimeout=redisIdleTimeout

	seckillConf.RedisProxy2LayerConf.RedisMaxIdle=redisProxy2LayerMaxIdle
	seckillConf.RedisProxy2LayerConf.RedisMaxActive=redisProxy2LayerMaxActive
	seckillConf.RedisProxy2LayerConf.RedisIdleTimeout=redisProxy2LayerIdleTimeout
	writeGoNums,err := beego.AppConfig.Int("write_proxy2layer_goroutine_num")
	if err!=nil{
		err = fmt.Errorf("init config failed,write_proxy2layer_goroutine_num:%v",err)
		return
	}
	seckillConf.WriteProxy2LayerGoroutineNum=writeGoNums

	readGoNums,err := beego.AppConfig.Int("read_layer2proxy_goroutine_num")
	if err!=nil{
		err = fmt.Errorf("init config failed,read_layer2proxy_goroutine_num:%v",err)
		return
	}
	seckillConf.ReadProxy2LayerGoroutineNum=readGoNums





	etcdtimeout ,err:=beego.AppConfig.Int("etcd_timeout")
	if err!=nil{
		err = fmt.Errorf("init config failed,read etcd timeout error")
		return
	}
	seckillConf.EtcdConf.Timeout=etcdtimeout
	seckillConf.EtcdConf.EtcdSecKeyPrefix = beego.AppConfig.String("etcd_sec_key_prefix")
	if len(seckillConf.EtcdConf.EtcdSecKeyPrefix)==0{
		err = fmt.Errorf("init config failed,read etcd_sec_key_prefix error%v",err)
		return
	}
	productKey := beego.AppConfig.String("etcd_product_key")
	if len(productKey)==0{
		err = fmt.Errorf("init config failed,read etcd_sec_key error%v",err)
		return
	}
	if strings.HasSuffix(seckillConf.EtcdConf.EtcdSecKeyPrefix,"/")== false{
		seckillConf.EtcdConf.EtcdSecKeyPrefix = seckillConf.EtcdConf.EtcdSecKeyPrefix+"/"
	}

	seckillConf.EtcdConf.EtcdSecProductKey =fmt.Sprintf("%s%s",seckillConf.EtcdConf.EtcdSecKeyPrefix,productKey)

	seckillConf.LogPath=beego.AppConfig.String("log_path")
	seckillConf.LogLevel=beego.AppConfig.String("log_level")
	seckillConf.CookieSecretKey=beego.AppConfig.String("cookie_secretkey")




	seckillConf.AccessLimitConf.UserSecAccessLimit,err=beego.AppConfig.Int("user_sec_access_limit")
	referList := beego.AppConfig.String("refer_whitelist")
	if len(referList) > 0 {
		seckillConf.ReferWhiteList = strings.Split(referList, ",")
	}

	ipLimit,err := beego.AppConfig.Int("ip_sec_access_limit")
	if err!=nil{
		err = fmt.Errorf("init config failed,read ip_sec_access_limit error %v",err)
		return
	}
	seckillConf.AccessLimitConf.IPSecAccessLimit = ipLimit



	minIdLimit, err := beego.AppConfig.Int("user_min_access_limit")
	if err != nil {
		err = fmt.Errorf("init config failed, read user_min_access_limit error:%v", err)
		return
	}
	seckillConf.AccessLimitConf.UserMinAccessLimit = minIdLimit

	minIpLimit, err := beego.AppConfig.Int("ip_min_access_limit")
	if err != nil {
		err = fmt.Errorf("init config failed, read ip_min_access_limit error:%v", err)
		return
	}
	seckillConf.AccessLimitConf.IPMinAccessLimit = minIpLimit

	return
}