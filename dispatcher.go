package main

import (
	"log"
	"fmt"
	"time"
	"context"
	"encoding/json"
	"gosearch/helper"
	"gosearch/facelib"
)

// 消息守候线程 -- 正常不会结束
func dispatcher(groupsList string) {
	log.Println("dispatcher() start")

	goroutineDelta <- +1
	defer func(){goroutineDelta <- -1}()

	// 读取特征数据
	facelib.ReadData(groupsList)

	// 初始化redis连接
	err := helper.InitRDB()
	if err!=nil {
		log.Println(err)
		return
	}

	// 注册消息队列
	pubsub := helper.Rdb.Subscribe(context.Background(), helper.REDIS_QUEUENAME)
	ch := pubsub.Channel()
	defer pubsub.Close()

	log.Println("rdb subscribed.")

	// 收取消息 - 一直循环
	for msg := range ch {
		log.Println(msg.Channel, len(msg.Payload))

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
	requestId, result, err := faceSearch(payload)
	if err!=nil {
		log.Println("f() Error: ", err)
		result = "{\"code\":-1}"
	}

	log.Printf("[%v] %s %s", time.Since(start), requestId, result)

	if requestId!="NO_RECIEVER" {
		// 返回结果
		err = helper.Rdb.Publish(context.Background(), "gosearch_"+requestId, result).Err()
		if err != nil {
			log.Println("f() Error: ", err)
		}
	}
}

func faceSearch(payload string) (string, string, error) {

	fields := make(map[string]interface{})
	if err := json.Unmarshal([]byte(payload), &fields); err != nil {
		return "", "", err
	}

	var requestId, groupId, action string
	var data []interface{}

	requestId, ok := fields["request_id"].(string)
	if !ok {
		return "", "", fmt.Errorf("need request_id")
	}

	groupId, ok = fields["group_id"].(string)
	if !ok {
		return requestId, "", fmt.Errorf("need group_id")
	}

	action, ok = fields["action"].(string)
	if !ok {
		return requestId, "", fmt.Errorf("need action")
	}

	//log.Println(reflect.TypeOf(fields["data"]))
	data, ok = fields["data"].([]interface{})
	if !ok {
		return requestId, "", fmt.Errorf("need data")
	}

	var testVec []float32 // 测试向量
	for _, item := range data {
		testVec = append(testVec, float32(item.(float64)))
	}

	var result []byte

	switch action {

	case "search": // 检索特征
		label, min := facelib.Search(groupId, testVec)
		log.Println(groupId, ThreshHold, label, min)

		if min < ThreshHold && label!="__BLANK__" { // __BLANK__ 说明特征已动态删除
			resultMap := map[string]interface{}{
				"label": label,
				"score": min,
				"code": 200,
			}
			result, _ = json.Marshal(resultMap)
		} else {
			result = []byte("{\"code\":0}")
		}

	case "add": // 新增特征值
		label, ok := fields["label"].(string)
		if !ok {
			return requestId, "", fmt.Errorf("need label")
		}
		facelib.AddNewData(groupId, label, testVec)
		result = []byte("{\"code\":200}")

	case "remove":  // 删除特征值：不真的删除，只删除labelname
		label, ok := fields["label"].(string)
		if !ok {
			return requestId, "", fmt.Errorf("need label")
		}
		facelib.RemoveData(groupId, label)
		result = []byte("{\"code\":200}")		

	default:
		log.Println("faceSearch() unknown action:", action)
		result = []byte("{\"code\":-2}")

	}

	return requestId, string(result), nil
}