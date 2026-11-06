package main

import (
	"fmt"
	jyj "jyj-auto/internal"
	"net/http"
)

func main() {
	port := "9102"

	fmt.Println("服务已启动：port:" + port)
	_ = http.ListenAndServe(":"+port, jyj.Handler())
}
