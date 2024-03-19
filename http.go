package main

import (
	"encoding/json"
	"net/http"
	"okx_data/common/logger"
	"strconv"
	"strings"
	"time"
)

func getSymbolVolatilityHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		symbol := r.URL.Query().Get("symbol")
		symbol = strings.ToUpper(symbol)

		calVolatilityMsStr := r.URL.Query().Get("volatilityTimeMs")
		calVolatilityMs, err := strconv.ParseInt(calVolatilityMsStr, 10, 64)
		if err != nil {
			logger.Info("volatilityTimeMs:%s is invalid", calVolatilityMsStr)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		volatility, ok := symbolVolatilities.GetVolatility(symbol)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var prices []float64
		cutoffTime := time.Now().Add(-time.Duration(calVolatilityMs) * time.Millisecond)
		for _, price := range volatility.Prices {
			if price.UpdateTime.After(cutoffTime) {
				prices = append(prices, price.Value) // 将价格添加到 relevantPrices 中
			}
		}
		response := struct {
			Code int `json:"code"`
			Data struct {
				Volatility float64   `json:"volatility"`
				Prices     []float64 `json:"prices"`
			} `json:"data"`
		}{
			Code: http.StatusOK,
		}

		response.Data.Volatility = volatility.Value
		response.Data.Prices = prices

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
