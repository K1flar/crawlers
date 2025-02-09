package main

import (
	"fmt"
	"log"
	"os"

	dotenv "github.com/joho/godotenv"
)

func main() {
	err := dotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("this is a cli")
	fmt.Println("args: ", os.Args)
}
