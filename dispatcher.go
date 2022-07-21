package main

import (
	"log"
	"fmt"
	"time"
	"context"
	"encoding/json"

	"antigen-go/helper"
	"antigen-go/gotf"
)

// 消息守候线程 -- 正常不会结束
func dispatcher(queueNum string) {
	log.Println("dispatcher() start")

	goroutineDelta <- +1
	defer func(){goroutineDelta <- -1}()

	// 注册消息队列
	pubsub := helper.Rdb.Subscribe(context.Background(), helper.REDIS_QUEUENAME+queueNum)
	ch := pubsub.Channel()
	defer pubsub.Close()

	log.Println("rdb subscribed -->", helper.REDIS_QUEUENAME+queueNum)

	// 收取消息 - 一直循环
	for msg := range ch {
		log.Printf("<-- %s [%d]", msg.Channel, len(msg.Payload))

		goroutineDelta <- +1
		go f(msg.Payload)		
	}

	log.Println("dispatcher() leave")
}

// 实际处理 gosearch
// payload 格式：
//	{ "request_id" : "", "data": [1, 2, 3, ...]}
func f(payload string) {
	defer func(){goroutineDelta <- -1}()

	start := time.Now()
	requestId, result, err := porcessApi(payload)
	if err!=nil {
		log.Println("f() Error: ", err)
		result = "{\"code\":-1}"
	}

	if requestId!="NO_RECIEVER" {
		// 返回结果
		err = helper.Rdb.Publish(context.Background(), requestId, result).Err()
		if err != nil {
			log.Println("f() Error: ", err)
		}

		log.Printf("--> %s [%d]", requestId, len(result))
	}

	log.Printf("[%v] %s", time.Since(start), requestId)
}

func porcessApi(payload string) (string, string, error) {
	retJson := map[string]interface{}{"code":0}

	fields := make(map[string]interface{})
	if err := json.Unmarshal([]byte(payload), &fields); err != nil {
		return "", "", err
	}

	//log.Println(fields)

	var requestId string

	requestId, ok := fields["request_id"].(string)
	if !ok {
		return "", "", fmt.Errorf("need request_id")
	}

	data, ok := fields["data"].(map[string]interface{})
	if !ok {
		return requestId, "", fmt.Errorf("need data")
	}

	//log.Println(data)

	var result []byte

	switch data["api"].(string) {
	case "bert_qa":
		ans, err := gotf.BertQA(data["corpus"].(string), data["question"].(string))
		if err!=nil {
			retJson["code"] = 9002
			retJson["msg"] = err.Error()
		} else {
			retJson["code"] = 0
			retJson["data"] = ans
		}

	default:
		log.Println("faceSearch() unknown api:", data["api"])
		result = []byte("{\"code\":-2}")
		retJson["code"] = 9001
		retJson["msg"] = "unknown api"
	}

	result, err := json.Marshal(retJson)
	if err != nil {
		return requestId, "", err
	}

	return requestId, string(result), nil
}