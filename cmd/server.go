package main

import (
	"ytsruh.com/envoy/pkg/cron"
	"ytsruh.com/envoy/pkg/database"
	"ytsruh.com/envoy/pkg/server"
	"ytsruh.com/envoy/pkg/utils"
)

func main() {
	env, err := utils.LoadAndValidateEnv()
	if err != nil {
		panic(err)
	}
	// Start a database service
	dbService, err := database.NewService(env.DB_PATH)
	if err != nil {
		panic(err)
	}

	// Start the cron service
	cronService := cron.New(dbService.GetDB())
	cronService.Start()

	server := server.New(":8080", dbService)
	server.Start()
}
