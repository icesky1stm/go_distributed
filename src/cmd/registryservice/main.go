package main

import (
	"context"
	"fmt"
	"go_distributed/src/registry"
	"log"
	"net/http"
)

func main() {

	registry.SetupRegistryService()

	// 设置http的服务注册
	http.Handle("/services", &registry.RegistryService{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var svr http.Server
	svr.Addr = registry.ServerPort

	// 运行服务
	go func() {
		log.Println(svr.ListenAndServe())
		cancel()
	}()

	// 关闭服务的指令
	go func() {
		fmt.Printf("启动----服务[注册中心], 按任意键关闭.\n")
		var s string
		fmt.Scanln(&s)
		ctx = context.WithValue(ctx, "stopKey", s)
		svr.Shutdown(ctx)
		cancel()
	}()

	// 获取结果
	<-ctx.Done()
	// 获取跨线程传递信息
	var ss string = ctx.Value("stopKey").(string)

	fmt.Printf("停止----服务[注册中心],收到字符[%s]\n", ss)

}
