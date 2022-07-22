package main

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"

	"antigen-go/go-infer/cli"
	"antigen-go/go-infer/types"
	"antigen-go/models"
)


var (
	rootCmd = &cobra.Command{
		Use:   "antigen-go",
		Short: "antigen to detect gen-test result",
	}
)

func init() {
	// 添加模型实例
	types.ModelList = append(types.ModelList, &models.BertQA{})
	types.ModelList = append(types.ModelList, &models.EchoModel{})

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
