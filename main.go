package main

import (
	"log"

	"json1/services"
)

func main() {
	service, err := services.NewJsonFileService("file.json")
	if err != nil {
		log.Fatal(err)
	}

	service.PrintUserGroup()
}
