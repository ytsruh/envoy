package main

import (
	"ytsruh.com/envman/pkg/routes"
	"ytsruh.com/envman/pkg/server"
	"ytsruh.com/envman/pkg/utils"
)

func main() {
	_, err := utils.LoadAndValidateEnv()
	if err != nil {
		panic(err)
	}
	routes := routes.RegisterRoutes()
	server := server.New(":8080", routes)
	server.Start()
}
