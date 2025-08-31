package main

import (
	"ytsruh.com/envoy/pkg/database"
	"ytsruh.com/envoy/pkg/server"
	"ytsruh.com/envoy/pkg/utils"
)

func main() {
	env, err := utils.LoadAndValidateEnv()
	if err != nil {
		panic(err)
	}
	dbService := database.NewService(env.DB_PATH)
	server := server.New(":8080", dbService)
	server.Start()
}
