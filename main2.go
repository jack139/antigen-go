package main

import (
	"log"
	"fmt"
	"os"
	"github.com/spf13/cobra"
	"gosearch/facelib"
)

var (
	ThreshHold = float32(1.7) // 阈值

	// Receives the change in the number of goroutines
	goroutineDelta = make(chan int)
	needToCreateANewGoroutine = bool(true)

	// cobra 命令行
	rootCmd = &cobra.Command{
		Use:   "gosearch",
		Short: "gosearch for yhfacelib",
	}

	// 注册数据来自数据库
	evalCmd = &cobra.Command{
		Use:   "eval <data file> <group id>",
		Short: "evaluation test",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("need <data file> and <group id>")
			}

			// 设置参数
			face, _ := cmd.Flags().GetUint16("face")
			gonum, _ := cmd.Flags().GetUint16("gonum")
			ThreshHold, _ = cmd.Flags().GetFloat32("threshold")
			facelib.LimitFace = int(face)
			facelib.GONUM = int(gonum)

			// 测试
			searchTest(args[1], args[0])
			return nil
		},
	}

	// 注册数据从文件读入
	eval2Cmd = &cobra.Command{
		Use:   "eval2 <reg data file> <test data file>",
		Short: "evaluation test by file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("need <reg data file> <test data file>")
			}

			// 设置参数
			face, _ := cmd.Flags().GetUint16("face")
			gonum, _ := cmd.Flags().GetUint16("gonum")
			ThreshHold, _ = cmd.Flags().GetFloat32("threshold")
			facelib.LimitFace = int(face)
			facelib.GONUM = int(gonum)

			// 测试
			searchTest2("eval2", args[0], args[1])
			return nil
		},
	}

	serverCmd = &cobra.Command{
		Use:   "server <group_id list>",
		Short: "start goserver service",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("group_id list needed")
			}

			// 设置参数
			face, _ := cmd.Flags().GetUint16("face")
			gonum, _ := cmd.Flags().GetUint16("gonum")
			ThreshHold, _ = cmd.Flags().GetFloat32("threshold")
			facelib.LimitFace = int(face)
			facelib.GONUM = int(gonum)

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

)

func init() {
	evalCmd.Flags().Float32P("threshold", "t", 1.7, "adjustment threshold")
	evalCmd.Flags().Uint16P("gonum", "n", 8, "goroutine num")
	evalCmd.Flags().Uint16P("face", "f", 3, "limit faces to register")
	rootCmd.AddCommand(evalCmd)

	eval2Cmd.Flags().Float32P("threshold", "t", 1.7, "adjustment threshold")
	eval2Cmd.Flags().Uint16P("gonum", "n", 8, "goroutine num")
	eval2Cmd.Flags().Uint16P("face", "f", 3, "limit faces to register")
	rootCmd.AddCommand(eval2Cmd)

	serverCmd.Flags().Float32P("threshold", "t", 1.7, "adjustment threshold")
	serverCmd.Flags().Uint16P("gonum", "n", 8, "goroutine num")
	serverCmd.Flags().Uint16P("face", "f", 3, "limit faces to register")
	rootCmd.AddCommand(serverCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
