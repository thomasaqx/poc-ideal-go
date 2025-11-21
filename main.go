package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/thomasaqx/poc-ideal-go.git/internal/client"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: impossible to load .env file. Using environment variables.")
	}

	apiKey := os.Getenv("YAHOO_API_KEY")
	if apiKey == "" {
		log.Fatal("Erro: the environment variable YAHOO_API_KEY is not set.")
	}

	fmt.Println("API Key loaded successfully")

	// Create Yahoo Finance client
	yahooClient := client.NewYahooClient(apiKey)

	// Test: fetch information for Apple (AAPL)
	fmt.Println("\nFetching information for stock AAPL...")
	quote, err := yahooClient.GetQuote("AAPL")
	if err != nil {
		log.Printf("Error fetching quote: %v\n", err)
	} else {
		if len(quote.QuoteResponse.Result) > 0 {
			result := quote.QuoteResponse.Result[0]
			fmt.Printf("\nSSymbol: %s\n", result.Symbol)
			fmt.Printf("Name: %s\n", result.LongName)
			fmt.Printf("Price: %.2f %s\n", result.RegularMarketPrice, result.Currency)
			fmt.Printf("Change: %.2f%%\n", result.RegularMarketChangePercent)
		}
	}
}
