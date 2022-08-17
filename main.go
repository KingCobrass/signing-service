package main

import (
	"fmt"
	"log"
	"runtime"
	"signing-service/api"
	"signing-service/config"
)

const (
// ListenAddress =
// TODO: add further configuration parameters here ...
)

func main() {

	log.Printf("CPU Cores available in machine: %v\n", runtime.NumCPU())

	server := api.NewServer(fmt.Sprintf(":%s", config.Env.Port))

	if err := server.Run(); err != nil {
		log.Fatal("Could not start server on ", config.Env.Port)
	}
}
