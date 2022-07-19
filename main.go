package main

import (
	"log"
	"fmt"
	"os"
	"path"
	"errors"
	"github.com/spf13/cobra"
	//"gosearch/facelib"

	"antigen-go/http"
	"antigen-go/gotf"
)

/* 预训练模型路径 */
const(
	modelPath = "saved-model"
	vocabPath = "saved-model/vocab_chinese.txt"
)

var (
	ThreshHold = float32(1.7) // 阈值

	// Receives the change in the number of goroutines
	goroutineDelta = make(chan int)
	needToCreateANewGoroutine = bool(true)

	// cobra 命令行
	rootCmd = &cobra.Command{
		Use:   "antigen-go",
		Short: "antigen to detect gen-test result",
	}

	// http 服务
	httpCmd = &cobra.Command{
		Use:   "http <port>",
		Short: "start http service",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("need port number")
			}

			// 保存 cmd
			//httphelper.HttpCmd = cmd

			// 启动 http 服务
			http.RunServer(args[0])
			// 不会返回
			return nil
		},
	}

	// Dispatcher
	serverCmd = &cobra.Command{
		Use:   "server <model path>",
		Short: "start dispatcher service",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("model path needed")
			}

			/* 初始化模型 */
			gotf.InitModel(path.Join(args[0], modelPath), path.Join(args[0], vocabPath))

			// 启动 分发服务
			go dispatcher()

			numGoroutines := 0
			for diff := range goroutineDelta {
				numGoroutines += diff
				log.Printf("Goroutines = %d\n", numGoroutines)
				if numGoroutines == 0 { os.Exit(0) }
			}
			return nil
		},
	}

)

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	
	rootCmd.AddCommand(httpCmd)
	rootCmd.AddCommand(serverCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
