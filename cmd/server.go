//go:generate templ generate -path ../server/views
package main

import (
	"ytsruh.com/envoy/server"
	"ytsruh.com/envoy/server/utils"
)

func main() {
	env, err := utils.LoadAndValidateEnv()
	if err != nil {
		panic(err)
	}

	s := server.New(":8080", env)

	s.Start()
}
