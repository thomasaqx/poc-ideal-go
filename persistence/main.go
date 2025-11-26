package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
)

type Quote struct {
	Symbol               string  `json:"symbol"`
	RegularMarketPrice   float64 `json:"regularMarketPrice"`
	RegularMarketDayHigh float64 `json:"regularMarketDayHigh"`
	RegularMarketDayLow  float64 `json:"regularMarketDayLow"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: impossible to load .env file. Using environment variables.")
	}

	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		log.Fatal("Error : the environment variable MYSQL_DSN is not set.")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Error connecting to the database:", err)
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    "watchlist-topic",
		GroupID:  "persistence-group",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	defer r.Close()

	log.Println("Persistence Service started (JSON). Waiting...")

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading Kafka: %v", err)
			continue
		}

		// Try to unmarshal the received JSON into an object
		var quote Quote
		err = json.Unmarshal(m.Value, &quote)
		if err != nil {
			// If there is an error here, it is likely an old message
			log.Printf("Ignoring invalid message: %s", string(m.Value))
			continue
		}

		fmt.Printf("Processing: %s | Price: %.2f\n", quote.Symbol, quote.RegularMarketPrice)

		err = saveToDB(db, quote.Symbol)
		if err != nil {
			log.Printf("Error saving %s: %v", quote.Symbol, err)
		} else {
			log.Printf("%s saved to the database!", quote.Symbol)
		}
	}
}

func saveToDB(db *sql.DB, symbol string) error {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	// INSERT IGNORE to avoid duplicates
	query := "INSERT IGNORE INTO watchlist (symbol) VALUES (?)"
	_, err := db.Exec(query, symbol)
	return err
}
