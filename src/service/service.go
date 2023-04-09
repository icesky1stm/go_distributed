package service

// 用来注册和启动service

import (
	"context"
	"fmt"
	"go_distributed/src/registry"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Start(ctx context.Context, host, port string,
	reg registry.Registration,
	registerHandlersFunc func()) (context.Context, error) {

	//设置绑定
	registerHandlersFunc()
	//启动服务
	ctx = startService(ctx, reg, host, port)

	// 注册服务
	err := registry.RegisterService(reg)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func startService(ctx context.Context, reg registry.Registration,
	host, port string) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	var svr http.Server
	svr.Addr = ":" + port

	go func() {
		log.Println(svr.ListenAndServe())
		// 取消注册
		err := registry.ShutdownService(reg.ServiceURL)
		if err != nil {
			log.Println(err)
		}
		cancel()
	}()

	go func() {
		fmt.Printf("%v startd, Press any key to stop.\n", reg.ServiceName)
		var s string
		fmt.Scanln(&s)

		// 取消注册
		err := registry.ShutdownService(reg.ServiceURL)
		if err != nil {
			log.Println(err)
		}
		svr.Shutdown(ctx)
		cancel()
	}()

	// 增加对signal信号的获取,优雅退出
	go func() {
		fmt.Println("开始等待系统syscall信号...")
		c := make(chan os.Signal)
		// SIGHUP: terminal closed
		// SIGINT: Ctrl+C
		// SIGTERM: program exit
		// SIGQUIT: Ctrl+/
		signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		// 阻塞，直到接受到退出信号，才停止进程
		for i := range c {
			switch i {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				// 这里做一些清理操作或者输出相关说明，比如 断开数据库连接
				fmt.Println("receive exit signal [", i.String(), "],准备退出exit...")
				// 取消注册
				err := registry.ShutdownService(reg.ServiceURL)
				if err != nil {
					log.Println(err)
				}
				svr.Shutdown(ctx)
				cancel()
			}
		}
	}()

	return ctx
}
