package db2struct

import (
	"flag"
	"fmt"
	"os/exec"
	"testing"
)

func TestDB2Struct(t *testing.T) {

	flag.Parse()

	if len(host) == 0 || len(port) == 0 || len(username) == 0 || len(pwd) == 0 || len(database) == 0 {
		return
	}

	//initVar()

	//	convert aa,bb to 'aa','bb'
	tables = convtables(tables) // 转换

	fmt.Println("正在启动生成结构体...")

	execGenStruct()

	var err error
	fmt.Println("去除无用的引用包...")
	cmd := exec.Command("go imports", "-w", path)
	if err = cmd.Run(); err != nil {
		fmt.Println(err)
	}

	// format
	fmt.Println("格式化文件...")
	cmdFmt := exec.Command("gofmt", "-l", "-w", path)
	if err = cmdFmt.Run(); err != nil {
		fmt.Println(err)
	}
	fmt.Println("结构体生成完成")
}
