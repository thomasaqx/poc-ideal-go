package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/thomasaqx/poc-ideal-go/internal/client"
	"github.com/thomasaqx/poc-ideal-go/internal/queue"
	"github.com/thomasaqx/poc-ideal-go/internal/storage"
	pb "github.com/thomasaqx/poc-ideal-go/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCServer exposes the gRPC handlers
type GRPCServer struct {
	pb.UnimplementedAssetServiceServer
	yahoo     *client.YahooClient
	watchlist *storage.Watchlist
	producer  *queue.KafkaProducer
}

// NewGRPCServer wires every dependency required by the gRPC service.
func NewGRPCServer(yahoo *client.YahooClient, watchlist *storage.Watchlist, producer *queue.KafkaProducer) *GRPCServer {
	return &GRPCServer{
		yahoo:     yahoo,
		watchlist: watchlist,
		producer:  producer,
	}
}

// GetAssetPrice mirrors the HTTP handler that fetches data from Yahoo Finance.
func (s *GRPCServer) GetAssetPrice(ctx context.Context, req *pb.AssetRequest) (*pb.AssetPriceResponse, error) {
	symbol := sanitizeSymbol(req.GetSymbol())
	if symbol == "" {
		return nil, status.Error(codes.InvalidArgument, "symbol is required")
	}

	quoteResponse, err := s.yahoo.GetQuote(symbol)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch quote: %v", err)
	}
	if quoteResponse == nil || len(quoteResponse.QuoteResponse.Result) == 0 {
		return nil, status.Error(codes.NotFound, "symbol not found")
	}

	result := quoteResponse.QuoteResponse.Result[0]
	return &pb.AssetPriceResponse{
		Symbol: result.Symbol,
		Price:  result.RegularMarketPrice,
	}, nil
}

// AddAssetToWatchlist validates the symbol, fetches the latest quote, and enqueues it into Kafka.
func (s *GRPCServer) AddAssetToWatchlist(ctx context.Context, req *pb.AssetRequest) (*pb.SuccessResponse, error) {
	symbol := sanitizeSymbol(req.GetSymbol())
	if symbol == "" {
		return nil, status.Error(codes.InvalidArgument, "symbol is required")
	}

	quoteResponse, err := s.yahoo.GetQuote(symbol)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch quote: %v", err)
	}
	if quoteResponse == nil || len(quoteResponse.QuoteResponse.Result) == 0 {
		return nil, status.Error(codes.NotFound, "symbol not found")
	}

	quote := quoteResponse.QuoteResponse.Result[0]
	payload, err := json.Marshal(quote)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to encode quote: %v", err)
	}

	if err := s.producer.Publish(string(payload)); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish to kafka: %v", err)
	}

	return &pb.SuccessResponse{
		Message: fmt.Sprintf("asset %s sent to queue", quote.Symbol),
	}, nil
}

// GetWatchlist returns every symbol currently stored in MySQL.
func (s *GRPCServer) GetWatchlist(ctx context.Context, req *pb.Empty) (*pb.WatchlistResponse, error) {
	symbols, err := s.watchlist.GetAll()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch watchlist: %v", err)
	}

	return &pb.WatchlistResponse{Symbols: symbols}, nil
}

func sanitizeSymbol(symbol string) string {
	return strings.ToUpper(strings.TrimSpace(symbol))
}
