package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/thomasaqx/poc-ideal-go/internal/client"
	"github.com/thomasaqx/poc-ideal-go/internal/models"
	"github.com/thomasaqx/poc-ideal-go/internal/queue"
	"github.com/thomasaqx/poc-ideal-go/internal/storage"
)

var errSymbolNotFound = errors.New("symbol not found")

type application struct {
	apiKey    string
	client    *client.YahooClient
	watchlist *storage.Watchlist
	producer  *queue.KafkaProducer
}

func (app *application) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/assets", func(r chi.Router) {
		r.Get("/watchlist", app.handleGetWatchlist)
		r.Post("/watchlist/{symbol}", app.handleAddAssetToWatchlist)
		r.Get("/{symbol}", app.handleGetAssetPrice)
	})

	return r
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: impossible to load .env file. Using environment variables.")
	}

	apiKey := os.Getenv("YAHOO_API_KEY")
	if apiKey == "" {
		log.Fatal("Erro: the environment variable YAHOO_API_KEY is not set.")
	}

	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		log.Fatal("Erro: the environment variable MYSQL_DSN is not set.")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to open database connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	maxOpenConns := getEnvInt("MYSQL_MAX_OPEN_CONNS", 10)
	maxIdleConns := getEnvInt("MYSQL_MAX_IDLE_CONNS", 5)
	connMaxLifetime := getEnvDuration("MYSQL_CONN_MAX_LIFETIME", time.Hour)

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	kafkaBroker := getEnvString("KAFKA_BROKER", "127.0.0.1:9092")
	kafkaTopic := getEnvString("KAFKA_TOPIC", "watchlist-topic")
	kafkaProducer := queue.NewKafkaProducer(kafkaBroker, kafkaTopic)
	defer kafkaProducer.Close()

	app := &application{
		apiKey:    apiKey,
		client:    client.NewYahooClient(apiKey),
		watchlist: storage.NewWatchlist(db),
		producer:  kafkaProducer,
	}

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Server start in http://localhost:8080...")
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func (app *application) handleGetAssetPrice(w http.ResponseWriter, r *http.Request) {
	symbol := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "symbol")))
	if symbol == "" {
		errorJSON(w, http.StatusBadRequest, "symbol is required")
		return
	}

	assetPrice, err := app.getAssetPriceFromAPI(r.Context(), symbol)
	if err != nil {
		if errors.Is(err, errSymbolNotFound) {
			errorJSON(w, http.StatusNotFound, fmt.Sprintf("symbol %s not found", symbol))
			return
		}
		log.Printf("error fetching asset price for %s: %v", symbol, err)
		errorJSON(w, http.StatusBadGateway, "unable to fetch asset price")
		return
	}

	writeJSON(w, http.StatusOK, assetPrice)
}

func (app *application) handleGetWatchlist(w http.ResponseWriter, r *http.Request) {
	symbols, err := app.watchlist.GetAll()
	if err != nil {
		log.Printf("error fetching watchlist: %v", err)
		errorJSON(w, http.StatusInternalServerError, "unable to fetch watchlist")
		return
	}
	writeJSON(w, http.StatusOK, map[string][]string{"symbols": symbols})
}

func (app *application) getAssetPriceFromAPI(ctx context.Context, symbol string) (*models.AssetPrice, error) {
	// Context is currently unused but kept for future cancellation support.
	_ = ctx

	quote, err := app.client.GetQuote(symbol)
	if err != nil {
		return nil, err
	}

	if quote == nil || len(quote.QuoteResponse.Result) == 0 {
		return nil, fmt.Errorf("%w: %s", errSymbolNotFound, symbol)
	}

	result := quote.QuoteResponse.Result[0]
	return &models.AssetPrice{
		Symbol: symbol,
		Price:  result.RegularMarketPrice,
	}, nil
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("error encoding json response: %v", err)
	}
}

func errorJSON(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 0 {
		log.Printf("invalid value for %s (%s), using default %d", key, value, fallback)
		return fallback
	}
	return n
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	dur, err := time.ParseDuration(value)
	if err != nil {
		log.Printf("invalid duration for %s (%s), using default %s", key, value, fallback)
		return fallback
	}
	return dur
}

func getEnvString(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

//code review
func (app *application) handleAddAssetToWatchlist(w http.ResponseWriter, r *http.Request) {
	symbol := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "symbol")))
	if symbol == "" {
		errorJSON(w, http.StatusBadRequest, "symbol is required")
		return
	}

	// fetch data from Yahoo Finance API
	quoteResponse, err := app.client.GetQuote(symbol)
	if err != nil {
		log.Printf("error fetching quote for %s: %v", symbol, err)
		errorJSON(w, http.StatusBadGateway, "error fetching asset data")
		return
	}

	if quoteResponse == nil || len(quoteResponse.QuoteResponse.Result) == 0 {
		errorJSON(w, http.StatusNotFound, fmt.Sprintf("symbol %s not found in external API", symbol))
		return
	}

	quote := quoteResponse.QuoteResponse.Result[0]

	// marshal the quote to JSON
	jsonMessage, err := json.Marshal(quote)
	if err != nil {
		log.Printf("error marshalling quote to json: %v", err)
		errorJSON(w, http.StatusInternalServerError, "error processing data")
		return
	}

	//send json to kafka
	err = app.producer.Publish(string(jsonMessage))
	if err != nil {
		log.Printf("error publishing to kafka: %v", err)
		errorJSON(w, http.StatusInternalServerError, "error queueing request")
		return
	}

	//return success response
	writeJSON(w, http.StatusAccepted, map[string]string{
		"message": fmt.Sprintf("Asset %s (Price: %.2f) sent to queue", quote.Symbol, quote.RegularMarketPrice),
	})
}
