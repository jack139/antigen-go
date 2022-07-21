package api

import (
	"fmt"
	"log"
	"encoding/json"
	"antigen-go/helper"
)

/* api测试: bert_qa */
func ApiBertQA(reqData *map[string]interface{}) (*map[string]interface{}, error) {
	log.Println("ApiBertQA")

	// 检查参数
	corpus, ok := (*reqData)["corpus"].(string)
	if !ok {
		return &map[string]interface{}{"code":9101}, fmt.Errorf("need corpus")
	}

	question, ok := (*reqData)["question"].(string)
	if !ok {
		return &map[string]interface{}{"code":9102}, fmt.Errorf("need question")
	}

	// 构建reqData
	reqDataMap := map[string]interface{}{
		"api": "bert_qa",
		"corpus": corpus,
		"question": question,
	}

	requestId := helper.GenerateRequestId()


	// 注册消息队列，在发redis消息前注册, 防止消息漏掉
	pubsub := helper.Redis_subscribe(requestId)
	defer pubsub.Close()

	// 发 请求消息
	err := helper.Redis_publish_request(requestId, &reqDataMap)
	if err!=nil {
		return &map[string]interface{}{"code":9103}, err
	}

	// 收 结果消息
	respBytes := helper.Redis_sub_receive(pubsub)

	// 转换成map, 生成返回数据
	var respData map[string]interface{}

	if err := json.Unmarshal(respBytes, &respData); err != nil {
		return &map[string]interface{}{"code":9104}, err
	}

	// code==0 提交成功
	if respData["code"].(float64)!=0 { 
		return &map[string]interface{}{"code":int(respData["code"].(float64))}, fmt.Errorf(respData["msg"].(string))
	}

	// 返回区块id
	resp := map[string]interface{}{
		//"data" : respData["data"].(map[string]interface{}),  // data 数据
		"ans" : respData["data"].(string),  // data 数据
	}

	return &resp, nil
}
