package service

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"sync"
)



type SecLimitMgr struct {
	UserLimitMap map[int]*Limit
	IpLimitMap   map[string]*Limit
	lock         sync.Mutex
}



func antiSpam(req *SecRequest)(err error){
	//查看是否在id黑名单中
	_, ok := secKillConf.idBlackMap[req.UserId]
	if ok {
		err = fmt.Errorf("invalid request")
		logs.Error("useId[%v] is block by id black", req.UserId)
		return
	}
	//查看是否在ip黑名单中
	_, ok = secKillConf.ipBlackMap[req.ClientAddr]
	if ok {
		err = fmt.Errorf("invalid request")
		logs.Error("useId[%v] ip[%v] is block by ip black", req.UserId, req.ClientAddr)
		return
	}

	//查看访问次数限制
	//uid 频率控制
	secKillConf.secLimitMgr.lock.Lock()
	//uid 频率控制
	limit, ok := secKillConf.secLimitMgr.UserLimitMap[req.UserId]
	if !ok {
		limit = &Limit{
			secLimit: &SecLimit{},
			minLimit: &MinLimit{},
		}
		secKillConf.secLimitMgr.UserLimitMap[req.UserId] = limit
	}

	secIdCount := limit.secLimit.Count(req.AccessTime.Unix())
	minIdCount := limit.minLimit.Count(req.AccessTime.Unix())

	//ip 频率控制
	limit, ok = secKillConf.secLimitMgr.IpLimitMap[req.ClientAddr]
	if !ok {
		limit = &Limit{
			secLimit: &SecLimit{},
			minLimit: &MinLimit{},
		}
		secKillConf.secLimitMgr.IpLimitMap[req.ClientAddr] = limit
	}

	secIpCount := limit.secLimit.Count(req.AccessTime.Unix())
	minIpCount := limit.minLimit.Count(req.AccessTime.Unix())
	secKillConf.secLimitMgr.lock.Unlock()

	if secIpCount > secKillConf.AccessLimitConf.IPSecAccessLimit {
		err = fmt.Errorf("invalid request")
		logs.Error("invalid request IPSecAccessLimit")
		return
	}

	if minIpCount > secKillConf.AccessLimitConf.IPMinAccessLimit {
		err = fmt.Errorf("invalid request")
		logs.Error("invalid request IPMinAccessLimi")
		return
	}

	if secIdCount > secKillConf.AccessLimitConf.UserSecAccessLimit {
		err = fmt.Errorf("invalid request")
		logs.Error("invalid request UserSecAccessLimit")
		return
	}

	if minIdCount > secKillConf.AccessLimitConf.UserMinAccessLimit {
		err = fmt.Errorf("invalid request")
		logs.Error("invalid request UserMinAccessLimit")
		return
	}
	return
}








