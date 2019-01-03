package main

import (
	"context"
	"errors"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"log"
	"sync"
	"time"
)

var etcd *clientv3.Client
var waitGroup sync.WaitGroup
const KEY = "/test/mylock"

func main() {
	if err := initEtcd(); err != nil {
		return
	}
	defer etcd.Close()

	go testLock(1, time.Second * 5)
	go testLock(2, time.Second)

	waitGroup.Add(2)
	waitGroup.Wait()

	//if err := putValue(key,"hello-value"); err != nil {
	//	log.Fatal("添加值失败", err)
	//	return
	//}
	//log.Println("添加值成功")
	//
	//value, err := getValue(key)
	//if err != nil {
	//	log.Fatal("获取值失败", err)
	//	return
	//}
	//log.Println(value)
}

func initEtcd() error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"192.168.208.129:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal("etcd连接失败", err)
		return err
	}

	log.Println("etcd连接成功")
	etcd = cli
	return nil
}

func testLock(lockNumber int, processTime time.Duration) error {
	defer waitGroup.Done()

	ctx, cancel := context.WithTimeout(context.Background(), processTime + 5 * time.Second)
	session, err := concurrency.NewSession(etcd, concurrency.WithContext(ctx))
	if err != nil {
		log.Fatal("获取Session失败")
		return err
	}

	log.Println(lockNumber, " - 准备获取Lock")
	mutex := concurrency.NewMutex(session, KEY)
	if err := mutex.Lock(context.Background()); err != nil {
		log.Fatal("获取Lock失败")
		return err
	}
	log.Println(lockNumber, " - 获取Lock成功")

	// 模拟处理时间
	log.Println(lockNumber, " - 模拟处理时间 — ", processTime)
	time.Sleep(processTime)
	log.Println(lockNumber, " - 处理结束")

	cancel()
	if err := mutex.Unlock(context.Background()); err != nil {
		log.Fatal(lockNumber, " - 解除Lock失败")
		return err
	}

	log.Println(lockNumber, " - 解除Lock成功")
	return nil
}

func putValue(key, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := etcd.Put(ctx, key, value)
	cancel()
	if err != nil {
		return err
	}

	log.Println("添加值成功", resp)
	return nil
}

func getValue(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := etcd.Get(ctx, key, clientv3.WithPrefix())
	cancel()
	if err != nil {
		return "", err
	}

	log.Println("获取值成功", resp)

	if len(resp.Kvs) == 0 {
		return "", errors.New("值为空")
	}

	return string(resp.Kvs[0].Value), nil
}