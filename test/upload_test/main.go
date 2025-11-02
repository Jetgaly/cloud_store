package main

import (
	"bytes"
	"cloud_store/api/file"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"os"
	"strconv"

	"time"
)

var (
	filePath string = "/home/jjet/projects/GoProjects/cloud_store/test/file/bigImg.jpg"
	baseURL  string = "http://127.0.0.1:8080/api"
	token    string = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJuaWNrbmFtZSI6InRlc3QiLCJpc3MiOiJDbG91ZFN0b3JlIiwic3ViIjoidXNlci1hdXRoIiwiZXhwIjoxNzYyMDU4MjYxfQ.zYGGWU8eZcnbAFr5A5fWESaVQMz7DhT4qZk_LnDVcq4"
)

type HttpResp struct {
	Code string            `json:"Code"`
	Msg  string            `json:"Msg"`
	Data file.FileMetaInfo `json:"Data"`
}

// 流式计算
func CalculateSHA256Stream(filePath string) (string, int64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", 0, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	hash := sha256.New()
	// 获取文件信息
	fileInfo, e1 := file.Stat()
	if e1 != nil {
		return "", 0, fmt.Errorf("获取文件信息失败: %v", e1)
	}
	fileSize := fileInfo.Size()
	// 使用 io.Copy 自动处理缓冲区
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", 0, fmt.Errorf("计算哈希失败: %v", err)
	}
	return hex.EncodeToString(hash.Sum(nil)), fileSize, nil
}
func Init() *HttpResp {
	url := baseURL + "/file"
	hash, size, err := CalculateSHA256Stream(filePath)
	fmt.Println("hash:", hash)
	fmt.Println("size:", size)
	if err != nil {
		fmt.Println("CalculateSHA256Stream err:", err.Error())
		return nil
	}
	sizeStr := strconv.Itoa(int(size))
	req := file.UploadInitReq{
		Hash:     hash,
		Size:     sizeStr,
		FileName: "bigImg.jpg",
	}
	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("JSON 序列化失败: %v\n", err)
		return nil
	}

	// 创建请求
	request, e3 := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if e3 != nil {
		fmt.Printf("创建请求失败: %v\n", e3)
		return nil
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", token)
	request.Header.Set("User-Agent", "CloudStorage-Client/1.0")
	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(request)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	var data []byte = make([]byte, 1024)
	n, e := resp.Body.Read(data)
	if e != nil && e != io.EOF {
		fmt.Printf("请求失败: %v\n", e)
		return nil
	}
	d := data[:n]
	fmt.Println("d", string(d))
	var jsonReq HttpResp
	json.Unmarshal(d, &jsonReq)

	fmt.Println(jsonReq)
	return &jsonReq
	// 使用 io.ReadAll 读取完整响应
	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	fmt.Printf("读取响应失败: %v\n", err)
	// 	return
	// }

	// fmt.Printf("响应状态: %d\n", resp.StatusCode)
	// fmt.Printf("响应体: %s\n", string(body))
}
func upload(msg *HttpResp) error {
	url := baseURL + "/file/chunk"

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	cnt, _ := strconv.Atoi(msg.Data.ChunkCount)
	// 解析分片大小
	chunkSize, err := strconv.ParseInt(msg.Data.ChunkSize, 10, 64)
	if err != nil {
		return fmt.Errorf("解析分片大小失败: %v", err)
	}
	for i := 0; i < cnt; i++ {
		offset := int64(i) * chunkSize
		chunkData := make([]byte, chunkSize)
		n, err := file.ReadAt(chunkData, offset)
		if err != nil && err != io.EOF {
			return fmt.Errorf("读取分片 %d 失败: %v", i, err)
		}
		if n == 0 {
			return nil
		}
		chunkData = chunkData[:n]
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// 添加字段
		writer.WriteField("upid", msg.Data.UploadId)
		writer.WriteField("index", strconv.Itoa(i))
		part, err := writer.CreateFormFile("data", fmt.Sprintf("chunk_%d.dat", i))
		if err != nil {
			return fmt.Errorf("创建文件字段失败: %v", err)
		}
		if _, err := part.Write(chunkData); err != nil {
			return fmt.Errorf("写入文件数据失败: %v", err)
		}
		writer.Close()
		req, err := http.NewRequest("POST", url, body)
		if err != nil {
			return fmt.Errorf("创建请求失败: %v", err)
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", token)

		// 发送请求
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("请求失败: %v", err)
		}
		defer resp.Body.Close()
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("读取响应失败: %v", err)
		}
		var jsonReq HttpResp
		json.Unmarshal(respBody, &jsonReq)
		fmt.Println(jsonReq)

	}
	return nil
}
func finish(msg *HttpResp) {
	url := baseURL + "/file/finish"
	req := file.UploadFinishReq{
		UploadId: msg.Data.UploadId,
	}
	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("JSON 序列化失败: %v\n", err)
		return
	}

	// 创建请求
	request, e3 := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if e3 != nil {
		fmt.Printf("创建请求失败: %v\n", e3)
		return
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", token)
	request.Header.Set("User-Agent", "CloudStorage-Client/1.0")
	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(request)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var data []byte = make([]byte, 1024)
	n, e := resp.Body.Read(data)
	if e != nil && e != io.EOF {
		fmt.Printf("请求失败: %v\n", e)
		return
	}
	d := data[:n]
	fmt.Println("d", string(d))
	var jsonReq HttpResp
	json.Unmarshal(d, &jsonReq)

	fmt.Println(jsonReq)

}
func main() {
	msg := Init()
	upload(msg)
	finish(msg)
}
