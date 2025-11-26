package main

import (
	"log"
	config "yunyez/internal/common/config"
)

func main() {
	name := config.GetString("name")
	log.Printf("name: %s", name)
}