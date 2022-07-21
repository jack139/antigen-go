package main

import (
	"log"
	"fmt"
	"os"
	"errors"
	"strconv"
	"github.com/spf13/cobra"

	"antigen-go/http"
	//"antigen-go/gotf"
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

			// 启动 http 服务
			http.RunServer(args[0])

			return nil
		},
	}

	// Dispatcher
	serverCmd = &cobra.Command{
		Use:   "server <queue No.>",
		Short: "start dispatcher service",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("queue number needed")
			}

			_, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("queue number should be a integer")
			}

			// 启动 分发服务
			go dispatcher(args[0])

			numGoroutines := 0
			for diff := range goroutineDelta {
				numGoroutines += diff
				log.Printf("Goroutines = %d\n", numGoroutines)
				if numGoroutines == 0 { os.Exit(0) }
			}
			return nil
		},
	}

	/*
	// 测试
	testCmd = &cobra.Command{
		Use:   "test <model path>",
		Short: "test some",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) == 0 {
				return fmt.Errorf("model path needed")
			}

			ans, err := gotf.BertQA(
				"金字塔（英语：pyramid），在建筑学上是指锥体建筑物，著名的有埃及金字塔，还有玛雅卡斯蒂略金字塔、阿兹特克金字塔（太阳金字塔、月亮金字塔）等。", 
				"金字塔是什么？",
			)
			if err!=nil {
				log.Printf("ERROR: %s", err)
			} else {
				log.Printf("ans: %v", ans)
			}

			// 不会返回
			return nil
		},
	}
	*/
)

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(httpCmd)
	rootCmd.AddCommand(serverCmd)
	//rootCmd.AddCommand(testCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
