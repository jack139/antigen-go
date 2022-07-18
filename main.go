package main

import (
	"os"
	"path"

	"gotalk/bert"
	"gotalk/http"
)

/* 预训练模型路径 */
const(
	modelPath = "saved-model"
	vocabPath = "saved-model/vocab_chinese.txt"
)

/* 主入口 */
func main() {
	args := os.Args[1:]
	if len(args)==0 {
		panic("Need data path.")
	}

	/* 初始化模型 */
	bert.InitModel(path.Join(args[0], modelPath), path.Join(args[0], vocabPath))

	/* 启动server */
	http.RunServer()
}
