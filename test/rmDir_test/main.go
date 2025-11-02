package main

import (
	"fmt"
	"os"
)

func main() {
	err := os.RemoveAll("/home/jjet/projects/GoProjects/cloud_store/temp/1984834861734584320")
	if err != nil {
		fmt.Printf("删除文件夹失败: %v\n", err)
		return
	}
	fmt.Println("文件夹删除成功")
}
