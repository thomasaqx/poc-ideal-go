package client

// QuoteResponse represents the response from the quotes API
type QuoteResponse struct {
	QuoteResponse struct {
		Result []Quote `json:"result"`
		Error  *string `json:"error"`
	} `json:"quoteResponse"`
}

// Quote contains only the fields we expose to the API.
type Quote struct {
	Symbol               string  `json:"symbol"`
	RegularMarketPrice   float64 `json:"regularMarketPrice"`
	RegularMarketDayHigh float64 `json:"regularMarketDayHigh"`
	RegularMarketDayLow  float64 `json:"regularMarketDayLow"`
}

type HistoricalDataResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Currency             string   `json:"currency"`
				Symbol               string   `json:"symbol"`
				ExchangeName         string   `json:"exchangeName"`
				InstrumentType       string   `json:"instrumentType"`
				FirstTradeDate       int64    `json:"firstTradeDate"`
				RegularMarketTime    int64    `json:"regularMarketTime"`
				Gmtoffset            int      `json:"gmtoffset"`
				Timezone             string   `json:"timezone"`
				ExchangeTimezoneName string   `json:"exchangeTimezoneName"`
				RegularMarketPrice   float64  `json:"regularMarketPrice"`
				ChartPreviousClose   float64  `json:"chartPreviousClose"`
				PreviousClose        float64  `json:"previousClose"`
				Scale                int      `json:"scale"`
				PriceHint            int      `json:"priceHint"`
				CurrentTradingPeriod struct{} `json:"currentTradingPeriod"`
				TradingPeriods       [][]struct {
					Timezone  string `json:"timezone"`
					Start     int64  `json:"start"`
					End       int64  `json:"end"`
					Gmtoffset int    `json:"gmtoffset"`
				} `json:"tradingPeriods"`
				DataGranularity string   `json:"dataGranularity"`
				Range           string   `json:"range"`
				ValidRanges     []string `json:"validRanges"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					High   []float64 `json:"high"`
					Volume []int64   `json:"volume"`
					Low    []float64 `json:"low"`
					Open   []float64 `json:"open"`
					Close  []float64 `json:"close"`
				} `json:"quote"`
				Adjclose []struct {
					Adjclose []float64 `json:"adjclose"`
				} `json:"adjclose"`
			} `json:"indicators"`
		} `json:"result"`
		Error *string `json:"error"`
	} `json:"chart"`
}
