package http

import (
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"log"
	//"time"

	"antigen-go/helper"
	"antigen-go/api"
)


/* 入口 */
func RunServer(port string /*, userPath string*/) {

	/* router */
	r := router.New()
	r.GET("/", index)
	r.POST("/api/null", api.DoNonthing)
	r.POST("/api/test", api.ApiTest)

	log.Printf("start HTTP server at 0.0.0.0:%s\n", port)

	/* 启动server */
	s := &fasthttp.Server{
		Handler: helper.Combined(r.Handler),
		Name:    "FastHttpLogger",
	}
	log.Fatal(s.ListenAndServe(":" + port))
}

/* 根返回 */
func index(ctx *fasthttp.RequestCtx) {
	log.Printf("%v", ctx.RemoteAddr())
	ctx.WriteString("Hello world.")
}
