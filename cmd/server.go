//go:generate templ generate
package main

import (
	"log"

	"ytsruh.com/envoy/server"
	"ytsruh.com/envoy/server/cron"
	"ytsruh.com/envoy/server/database"
	"ytsruh.com/envoy/server/utils"
)

func main() {
	env, err := utils.LoadAndValidateEnv()
	if err != nil {
		panic(err)
	}

	dbService, err := database.NewService(env.DB_PATH)
	if err != nil {
		panic(err)
	}

	logger := log.New(log.Writer(), "", log.LstdFlags)
	cronService := cron.New(dbService.GetDB(), logger)
	cronService.AddJob("*/30 * * * * *", cron.DatabaseHealthCheck(dbService.GetDB(), logger))
	cronService.Start()

	srv, err := server.NewBuilder(":8080", dbService, env.JWT_SECRET).Build()
	if err != nil {
		log.Fatalf("failed to build server: %v", err)
	}

	srv.Start()
}
