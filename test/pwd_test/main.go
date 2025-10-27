package main

import (
	"cloud_store/utils"
	"fmt"
)

func main() {
	str := "123456"
	hash, _ := utils.HashPassword(str)
	fmt.Println("hash:" + hash)
	ok := utils.CheckPasswordHash(str, hash)
	fmt.Println(ok)
}
