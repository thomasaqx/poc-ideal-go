package models

type AssetPrice struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
}

type YahooResponse struct {
	QuoteResponse QuoteResponse `json:"quoteResponse"`
}

type QuoteResponse struct {
	Result []QuoteResult `json:"result"`
}

type QuoteResult struct {
	Symbol             string  `json:"symbol"`
	RegularMarketPrice float64 `json:"regularMarketPrice"` 
}
