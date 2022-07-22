package cli

import (
	"fmt"
	"errors"
	"strconv"
	"github.com/spf13/cobra"

	"antigen-go/go-infer/http"
	"antigen-go/go-infer/server"
	"antigen-go/go-infer/types"
	//"antigen-go/gotf"
)

var (
	// http 服务
	HttpCmd = &cobra.Command{
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

	// Dispatcher server
	ServerCmd = &cobra.Command{
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

			// 初始化模型
			for m := range types.ModelList {
				err = types.ModelList[m].Init()
				if err != nil {
					return fmt.Errorf("Init deep model fail: ", types.ModelList[m].ApiPath())
				}
			}

			// 启动 分发服务
			server.RunServer(args[0])

			return nil
		},
	}
)
