package http

import (
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"log"

	"antigen-go/go-infer/helper"
	"antigen-go/go-infer/types"
)


/* 入口 */
func RunServer(port string /*, userPath string*/) {

	/* router */
	r := router.New()
	r.GET("/", index)
	for path := range types.EntryMap {
		r.POST(path, apiEntry)
		log.Println("router added: ", path)
	}

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
