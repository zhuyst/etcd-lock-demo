package main

import (
	"context"
	"errors"
	"github.com/coreos/etcd/clientv3"
	"log"
	"time"
)

var etcd *clientv3.Client

func main() {
	if err := initEtcd(); err != nil {
		return
	}

	key := "/test/hello-key"

	if err := putValue(key,"hello-value"); err != nil {
		log.Fatal("添加值失败", err)
		return
	}
	log.Println("添加值成功")

	value, err := getValue(key)
	if err != nil {
		log.Fatal("获取值失败", err)
		return
	}
	log.Println(value)

	defer etcd.Close()
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