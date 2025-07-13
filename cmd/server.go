package main

import (
	"ytsruh.com/envman/pkg/database"
	"ytsruh.com/envman/pkg/routes"
	"ytsruh.com/envman/pkg/server"
	"ytsruh.com/envman/pkg/utils"
)

func main() {
	env, err := utils.LoadAndValidateEnv()
	if err != nil {
		panic(err)
	}
	db := database.New(env.TURSO_DATABASE_URL, env.TURSO_AUTH_TOKEN)
	router := server.NewRouter()
	routes.RegisterRoutes(router, db)
	server := server.New(":8080", router)
	server.Start()
}
