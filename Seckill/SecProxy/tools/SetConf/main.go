package main

import(
	"context"
	etcd_client "go.etcd.io/etcd/clientv3"
	"encoding/json"
	"fmt"
	"time"
)

type SecInfoConf struct{
	ProductId   int
	StartTime   int
	EndTime     int
	Status      int
	Total       int
	left        int
}

const(
	EtcdKey = "/zcz/secskill/product"
)


func SetLogConfToEtcd() {
	cli,err :=etcd_client .New(etcd_client.Config{
		Endpoints:            []string{"127.0.0.1:2379"},
		DialTimeout:          5*time.Second,
	})
	if err != nil{
		fmt.Println("connect failed,err",err)
	}
	fmt.Println("connect succ")
	defer cli.Close()

	var SecInfoConfArr []SecInfoConf

	SecInfoConfArr = append(
		SecInfoConfArr,
		SecInfoConf{
			ProductId:1022,
			StartTime:1574665514,
			EndTime:1574665614,
			Status:0,
			Total:10000,
			left:10000,
		},
	)
	SecInfoConfArr = append(
		SecInfoConfArr,
		SecInfoConf{
			ProductId:1052,
			StartTime:1517495965,
			EndTime:1517582365,
			Status:0,
			Total:900,
			left:900,
		},
	)

	SecInfoConfArr = append(
		SecInfoConfArr,
		SecInfoConf{
			ProductId:1062,
			StartTime:1572338046,
			EndTime:1572510846,
			Status:0,
			Total:900,
			left:900,
		},
	)

	data,err:=json.Marshal(SecInfoConfArr)
	if err!=nil{
		fmt.Println("json failed %v",err)
		return
	}
	ctx,cancel := context.WithTimeout(context.Background(),time.Second)
	_,err = cli.Put(ctx,EtcdKey,string(data))
	cancel()
	if err!=nil{
		fmt.Println("put failed%v",err)
		return
	}
	ctx,cancel = context.WithTimeout(context.Background(),time.Second)
	resp,err:=cli.Get(ctx,EtcdKey)
	cancel()
	if err!=nil{
		fmt.Println("get failed err",err)
	}
	for _,ev := range resp.Kvs{
		fmt.Printf("%s:%s\n",ev.Key,ev.Value)
	}

}


func main(){
	//etcd  读写操作
	SetLogConfToEtcd()
}