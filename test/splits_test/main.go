package main

import (
	"fmt"
	"strings"
)

func main() {
	str := "-"
	p := strings.Split(str, "-")
	fmt.Println(p)
	fmt.Println(len(p))
}
