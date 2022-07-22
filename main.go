package main

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"

	"antigen-go/go-infer/cli"
	"antigen-go/go-infer/types"

	"antigen-go/models/echo"
	//"antigen-go/models/qa"
	"antigen-go/models/embedding"
)


var (
	rootCmd = &cobra.Command{
		Use:   "antigen-go",
		Short: "antigen to detect gen-test result",
	}
)

func init() {
	// 添加模型实例
	//types.ModelList = append(types.ModelList, &qa.BertQA{})
	types.ModelList = append(types.ModelList, &echo.EchoModel{})
	types.ModelList = append(types.ModelList, &embedding.BertEMB{})

	// 添加 api 入口
	for m := range types.ModelList {
		types.EntryMap[types.ModelList[m].ApiPath()] = types.ModelList[m].ApiEntry
	}

	// 命令行设置
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(cli.HttpCmd)
	rootCmd.AddCommand(cli.ServerCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
