package main

import (
	"fmt"
	"log"
	"signing-service/api"
	"signing-service/config"
)

const (
// ListenAddress =
// TODO: add further configuration parameters here ...
)

func main() {

	server := api.NewServer(fmt.Sprintf(":%s", config.Env.Port))

	if err := server.Run(); err != nil {
		log.Fatal("Could not start server on ", config.Env.Port)
		log.Printf("Error: ", err.Error())
	}
}
