package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main(){

	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: impossible to load .env file. Using environment variables.")
	}

	apiKey := os.Getenv("YAHOO_API_KEY")
	if apiKey == "" {
		log.Fatal("Erro: the environment variable YAHOO_API_KEY is not set.")
	}

	fmt.Println("API Key loaded successfully")

}