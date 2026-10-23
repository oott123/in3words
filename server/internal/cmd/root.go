// Package cmd 定义命令行入口。
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewRootCommand 创建根命令。
func NewRootCommand() *cobra.Command {
	v := viper.New()
	var configFile string

	root := &cobra.Command{
		Use:           "in3words-server",
		Short:         "以 WordPress REST API 格式提供只读接口，直接读取 WordPress 的 MySQL 数据库",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringVarP(&configFile, "config", "c", "", "配置文件路径（支持 yaml/toml/json）")
	root.AddCommand(newServeCommand(v, &configFile))
	return root
}
