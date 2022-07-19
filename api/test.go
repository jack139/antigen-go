package api

import (
	"log"
	"github.com/valyala/fasthttp"

	"antigen-go/helper"
)

/* 空接口, 只进行签名校验 */
func DoNonthing(ctx *fasthttp.RequestCtx) {
	log.Println("doNonthing")

	// POST 的数据
	content := ctx.PostBody()

	// 验签
	data, err := helper.CheckSign(content)
	if err != nil {
		helper.RespError(ctx, 9000, err.Error())
		return
	}
	log.Printf("%v\n", *data)

	helper.RespJson(ctx, data) // 正常返回
	//helper.RespError(ctx, 99, "some errors") // 错误返回
}
