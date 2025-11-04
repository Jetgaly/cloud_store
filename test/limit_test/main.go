package main

import (
	"fmt"
	"io"
	"net/http"
)

var (
	url   = "http://127.0.0.1:8080/api/file/chunk"
	token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJuaWNrbmFtZSI6InRlc3QiLCJpc3MiOiJDbG91ZFN0b3JlIiwic3ViIjoidXNlci1hdXRoIiwiZXhwIjoxNzYyMjU5OTM1fQ.6aKwbPj3lPC5x0sAbGW7mXd0VR3YSdEz1-CSM9oEFdY"
)

func main() {
	req, e := http.NewRequest("POST", url, nil)
	if e != nil {
		fmt.Println("e", e.Error())
	}
	client := &http.Client{}
req.Header.Set("Authorization", token)

	for i := 0; i < 10; i++ {
		resp, errr := client.Do(req)
		//使用 io.ReadAll 读取完整响应
		if errr != nil {
			fmt.Println("errrr", errr.Error())
			return
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("读取响应失败: %v\n", err)
			return
		}
		fmt.Println(string(body))
	}
}
