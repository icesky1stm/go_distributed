package main

import (
	"context"
	"fmt"
	"go_distributed/src/log"
	"go_distributed/src/registry"
	"go_distributed/src/service"
	stlog "log"
)

func main() {
	log.Run("./distributed.log")
	host, port := "127.0.0.1", "4000"

	serviceAddress := fmt.Sprintf("http://%s:%s", host, port)

	r := registry.Registration{
		ServiceName:      registry.LogService,
		ServiceURL:       serviceAddress,
		RequiredServices: make([]registry.ServiceName, 0),
		ServiceUpdateURL: serviceAddress + "/services",
		HeartBeatURL:     serviceAddress + "/heartbeat",
	}

	ctx, err := service.Start(
		context.Background(),
		host,
		port,
		r,
		log.RegisterHandlers,
	)

	if err != nil {
		stlog.Fatalln("启动服务失败" + err.Error())
	}
	<-ctx.Done()

	fmt.Println("Shutting down logService.")

}
