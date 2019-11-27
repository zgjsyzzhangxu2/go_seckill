package service

import (
	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"
	"strconv"
	"time"
)
var (
	secKillConf *SecSkillConf
)

func InitService(serviceConf *SecSkillConf)(err error) {
	secKillConf = serviceConf
	//初始化黑名单(id ip)
	err = loadBlackList()
	if err!=nil{
		logs.Error("load blacklist err:%v",err)
		return
	}
	logs.Debug("init service succ,config:%v", secKillConf)

	err = initProxy2LayerRedis()
	if err!=nil{
		logs.Error("load proxy2Layer redis pool:%v",err)
		return
	}

	secKillConf.secLimitMgr = &SecLimitMgr{
		UserLimitMap: make(map[int]*Limit, 10000),
		IpLimitMap:   make(map[string]*Limit, 10000),
	}

	secKillConf.SecReqChan = make(chan *SecRequest, secKillConf.SecReqChanSize)
	secKillConf.UserConnMap = make(map[string]chan *SecResult, 10000)


	initRedisProcessFunc()
	if err!=nil{
		logs.Error("load pinitRedisProcessFunc:%v",err)
		return
	}
	return
}

func initRedisProcessFunc() {
	for i := 0; i < secKillConf.WriteProxy2LayerGoroutineNum; i++ {
		go WriteHandle()
	}

	for i := 0; i < secKillConf.ReadProxy2LayerGoroutineNum; i++ {
		go ReadHandle()
	}
}



func initProxy2LayerRedis()(err error) {
	secKillConf.proxy2LayerRedisPool = &redis.Pool{
		MaxIdle:   secKillConf.RedisProxy2LayerConf.RedisMaxIdle,
		MaxActive:   secKillConf.RedisProxy2LayerConf.RedisMaxActive,
		IdleTimeout: time.Duration(secKillConf.RedisProxy2LayerConf.RedisIdleTimeout)*time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", secKillConf.RedisProxy2LayerConf.RedisAddr)
		},
	}
	conn :=secKillConf.proxy2LayerRedisPool.Get()
	defer conn.Close()
	_, err =conn.Do("ping")
	if err !=nil{
		logs.Error("ping redis failed,err：%v",err)
		return
	}
	return

}

func loadBlackList()(err error) {
	secKillConf.ipBlackMap = make(map[string]bool, 10000)
	secKillConf.idBlackMap = make(map[int]bool, 10000)

	err = initBlackRedis()
	if err != nil {
		logs.Error("init black redis failed, err:%v", err)
		return
	}

	conn := secKillConf.blackRedisPool.Get()
	defer conn.Close()

	reply,err :=conn.Do("hgetall","idblacklist")

	idlist,err := redis.Strings(reply,err)
	if err!=nil{
		logs.Warn("hget all failed,err%v",err)
		return
	}
	for _,v :=range idlist{
		id ,err :=strconv.Atoi(v)
		if err !=nil{
			logs.Warn("invalid user id[%v]",id)
			continue
		}
		secKillConf.idBlackMap[id]=true
	}

	reply,err =conn.Do("hgetall","ipblacklist")
	iplist,err := redis.Strings(reply,err)
	if err!=nil{
		logs.Warn("hget all failed,err%v",err)
		return
	}
	for _,v :=range iplist{
		secKillConf.ipBlackMap[v]=true
	}
	go SyncIdBlackList()
	go SyncIpBlackList()

	return
}

func SyncIpBlackList(){
	var ipList []string
	lastTime :=time.Now().Unix()

	for{
		conn :=secKillConf.blackRedisPool.Get()

		reply,err :=conn.Do("BLPOP","blackiplist",time.Second)
		ip,err :=redis.String(reply,err)

		if err!=nil{
			conn.Close()
			continue
		}
		curTime :=time.Now().Unix()
		ipList = append(ipList,ip)
		if len(ipList)>100||curTime -lastTime>5 {
			secKillConf.RWBlackLock.Lock()
			for _,v :=range ipList{
				secKillConf.ipBlackMap[v] = true
			}
			secKillConf.RWBlackLock.Unlock()
			lastTime = curTime
			logs.Info("sync ip list from redis succ,ip[%v]",ipList)


		}

		conn.Close()


	}
}

func SyncIdBlackList(){
	//var idList []string
	//lastTime :=time.Now().Unix()

	for{
		conn :=secKillConf.blackRedisPool.Get()

		reply,err :=conn.Do("BLPOP","blackidlist",time.Second)
		id,err :=redis.Int(reply,err)

		if err!=nil{
			//logs.Debug("找不到")
			conn.Close()
			continue
		}
		secKillConf.RWBlackLock.Lock()
		secKillConf.idBlackMap[id] = true
		secKillConf.RWBlackLock.Unlock()

		logs.Info("sync id list from redis succ, ip[%v]", id)
		//curTime :=time.Now().Unix()
		//idList = append(idList,id)
		//if len(idList)>100||curTime -lastTime>5 {
		//	secKillConf.RWBlackLock.Lock()
		//	for _,v :=range idList{
		//		id ,err :=strconv.Atoi(v)
		//		if err !=nil{
		//			logs.Warn("invalid user id[%v]",id)
		//			continue
		//		}
		//		secKillConf.idBlackMap[id] = true
		//	}
		//	secKillConf.RWBlackLock.Unlock()
		//	lastTime = curTime
		//	logs.Info("sync ip list from redis succ,ip[%v]",idList)

		//}

		conn.Close()


	}
}


func initBlackRedis()(err error){
	secKillConf.blackRedisPool = &redis.Pool{
		MaxIdle:   secKillConf.RedisBlackConf.RedisMaxIdle,
		MaxActive:   secKillConf.RedisBlackConf.RedisMaxActive,
		IdleTimeout: time.Duration(secKillConf.RedisBlackConf.RedisIdleTimeout)*time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", secKillConf.RedisBlackConf.RedisAddr)
		},
	}
	conn :=secKillConf.blackRedisPool.Get()
	defer conn.Close()
	_, err =conn.Do("ping")
	if err !=nil{
		logs.Error("ping redis failed,err：%v",err)
		return
	}
	return

}
