package main

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"

	//"antigen-go/go-infer/types"
	"antigen-go/go-infer/cli"

	"antigen-go/gotf"
)


var (
	rootCmd = &cobra.Command{
		Use:   "antigen-go",
		Short: "antigen to detect gen-test result",
	}
)

/*  定义模型相关参数和方法  */
type MyModel struct{}

func (m MyModel) Init() error {
	return gotf.InitModel()
}

///////////////////////

func init() {
	cli.AModel = MyModel{}

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
