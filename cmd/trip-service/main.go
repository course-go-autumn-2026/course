package main

import (
	"log"

	"github.com/DaleCoopTP/grishin_lab1/internal/config"
)

func main() {
	if _, err := config.Load(); err != nil {
		log.Fatal(err)
	}
}
