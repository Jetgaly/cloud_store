package main

import (
	_ "cloud_store/core"
	"cloud_store/global"
	gcrontask "cloud_store/global/gCronTask"
	"cloud_store/router"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	//"time"
	//ginzap "github.com/gin-contrib/zap"
)

func main() {
	router.InitRouter()

	srv := &http.Server{
		Addr:    global.Config.ServeAt,
		Handler: global.Engine,
	}

	// 启动服务
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			global.Logger.Fatal(fmt.Sprintf("listen: %s\n", err))
		}
	}()

	global.Logger.Info("server running at " + global.Config.ServeAt)
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("closing server...")

	// 创建带超时的context，确保一定时间内强制退出
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.Shutdown(ctx); err != nil {
			global.Logger.Error("Server forced to shutdown:" + err.Error())
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		gcrontask.CleanTask.Stop()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		gcrontask.OSSUploadTask.Stop()
	}()
	wg.Wait()
	global.Logger.Sync()

	fmt.Println("exit gracefully")
}
