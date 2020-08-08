package main

import (
	"esmeapi/api"
	"fmt"
)

const appVersion = "0.0.1"

func main() {
	s := api.Server{}
	if err := s.Run(); err != nil {
		panic(fmt.Sprintf("cannot run server due to: %v", err))
	}
}
