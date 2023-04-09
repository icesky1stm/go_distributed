package main

import (
	"context"
	"fmt"
	"go_distributed/src/grades"
	"go_distributed/src/log"
	"go_distributed/src/registry"
	"go_distributed/src/service"
	stlog "log"
)

func main() {
	host, port := "127.0.0.1", "6000"

	serviceAddress := fmt.Sprintf("http://%s:%s", host, port)

	r := registry.Registration{
		ServiceName:      registry.GradingService,
		ServiceURL:       serviceAddress,
		RequiredServices: []registry.ServiceName{registry.LogService},
		ServiceUpdateURL: serviceAddress + "/services",
		HeartBeatURL:     serviceAddress + "/heartbeat",
	}

	ctx, err := service.Start(
		context.Background(),
		host,
		port,
		r,
		grades.RegisterHandlers,
	)
	if err != nil {
		stlog.Fatalln("启动服务失败" + err.Error())
	}

	logProvider, err := registry.GetProvider(registry.LogService)
	if err == nil {
		fmt.Printf("Logging service found at %s]\n", logProvider)
		log.SetClientLogger(logProvider, r.ServiceName)
	}
	<-ctx.Done()

	fmt.Println("Shutting down logService.")

}
