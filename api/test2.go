package api

import (
	//"log"
	"encoding/json"
	"github.com/valyala/fasthttp"

	"antigen-go/helper"
)

/* http测试 */
func ApiTest(ctx *fasthttp.RequestCtx) {
	//log.Println("APITest")

	// POST 的数据
	content := ctx.PostBody()

	// 验签
	reqData, err := helper.CheckSign(content)
	if err != nil {
		helper.RespError(ctx, 9000, err.Error())
		return
	}

	// 检查参数
	corpus, ok := (*reqData)["corpus"].(string)
	if !ok {
		helper.RespError(ctx, 9101, "need corpus")
		return
	}

	question, ok := (*reqData)["question"].(string)
	if !ok {
		helper.RespError(ctx, 9102, "need question")
		return
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
	err = helper.Redis_publish_request(requestId, &reqDataMap)
	if err!=nil {
		helper.RespError(ctx, 9103, err.Error())
		return		
	}

	// 收 结果消息
	respBytes := helper.Redis_sub_receive(pubsub)

	// 转换成map, 生成返回数据
	var respData map[string]interface{}

	if err := json.Unmarshal(respBytes, &respData); err != nil {
		helper.RespError(ctx, 9012, err.Error())
		return
	}

	// code==0 提交成功
	if respData["code"].(float64)!=0 { 
		helper.RespError(ctx, int(respData["code"].(float64)), respData["msg"].(string))  ///  提交失败
		return
	}

	// 返回区块id
	resp := map[string]interface{}{
		//"data" : respData["data"].(map[string]interface{}),  // data 数据
		"ans" : respData["data"].(string),  // data 数据
	}

	helper.RespJson(ctx, &resp)
}
