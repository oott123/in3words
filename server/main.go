// Command in3words-server 以 WordPress REST API 的格式提供只读接口。
package main

import (
	"context"
	"fmt"
	"os"

	_ "time/tzdata" // 精简镜像中可能没有时区数据库，timezone_string 需要它

	"github.com/oott123/in3words/server/internal/cmd"
)

func main() {
	if err := cmd.NewRootCommand().ExecuteContext(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
