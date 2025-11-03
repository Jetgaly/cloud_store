package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

var (
	baseURL string = "http://127.0.0.1:8080/api"
	token   string = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJuaWNrbmFtZSI6InRlc3QiLCJpc3MiOiJDbG91ZFN0b3JlIiwic3ViIjoidXNlci1hdXRoIiwiZXhwIjoxNzYyMTc2Mjk0fQ.BeedCR9CKgmgaMFzD3k-bQAo28y2I_EfJySLADdeMxY"
)

func main() {
	url := baseURL + "/file/1"
	out, err := os.Create("./tmp.bin")
	if err != nil {
		fmt.Println("创建文件失败:", err)
		return
	}
	defer out.Close()

	var start int64 = 0
	const chunkSize int64 = 1024 * 1024 // 1MB
	for {

		end := start + chunkSize - 1
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Println("init err", err.Error())
			return
		}
		rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)
		req.Header.Set("Authorization", token)
		req.Header.Set("Range", rangeHeader)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("请求失败:", err)
			return
		}
		if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
			fmt.Println("服务端返回错误:", resp.Status)
			return
		}
		// 写入文件
		n, err := io.Copy(out, resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Println("写入失败:", err)
			return
		}
		fmt.Println("s-e", start, end)

		// 已经写满本次 chunk，移动 offset
		start += n

		// 如果本次返回的内容小于 chunk，说明到尾了
		if n < chunkSize {
			fmt.Println("下载完成")
			break
		}
	}

	/*
		url := baseURL + "/file/1"
		out, err := os.Create("./tmp.bin")
		if err != nil {
			fmt.Println("创建文件失败:", err)
			return
		}
		defer out.Close()

		//var start int64 = 0
		//const chunkSize int64 = 1024 * 1024 // 1MB

		//end := start + chunkSize - 1
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Println("init err", err.Error())
			return
		}
		//rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)
		req.Header.Set("Authorization", token)
		//req.Header.Set("Range", rangeHeader)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("请求失败:", err)
			return
		}
		if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
			fmt.Println("服务端返回错误:", resp.Status)
			return
		}
		// 写入文件
		_, err = io.Copy(out, resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Println("写入失败:", err)
			return
		}
	*/
}
