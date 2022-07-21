package http

import (
	"log"
	"github.com/valyala/fasthttp"

	"antigen-go/go-infer/helper"
	"antigen-go/api"
)

// 处理函数入口类型
type entry func(*map[string]interface{}) (*map[string]interface{}, error)

var (
	// api 入口 与 处理过程 映射
	ENTRY_MAP = map[string]entry{
		"/api/echo"    : api.ApiNothing,
		"/api/bert_qa" : api.ApiBertQA,
	}
)

/* 空接口, 只进行签名校验 */
func apiEntry(ctx *fasthttp.RequestCtx) {
	log.Println("apiEntry")

	// POST 的数据
	content := ctx.PostBody()

	// 验签
	data, err := helper.CheckSign(content)
	if err != nil {
		helper.RespError(ctx, 9000, err.Error())
		return
	}

	for path := range ENTRY_MAP {
		if path == string(ctx.Path()) {
			ret, err := (ENTRY_MAP[path])(data)
			if err==nil {
				helper.RespJson(ctx, ret) // 正常返回
			} else {
				helper.RespError(ctx, (*ret)["code"].(int), err.Error()) 
			}
			return
		}
	}

	helper.RespError(ctx, 9900, "unknow path") 
}
