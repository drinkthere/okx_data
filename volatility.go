package main

import (
	"okx_data/common/logger"
	"okx_data/model"
	"sort"
	"sync"
	"time"
)

type Volatility struct {
	Value  float64
	Prices []model.Price
}

func NewVolatility() *Volatility {
	return &Volatility{}
}

func (v *Volatility) CalculateVolatility(calVolatilityMs int64) float64 {
	cutoffTime := time.Now().Add(-time.Duration(calVolatilityMs) * time.Millisecond)

	// 找到第一个不满足要求的时间点的索引
	index := sort.Search(len(v.Prices), func(i int) bool {
		return v.Prices[i].UpdateTime.After(cutoffTime)
	})
	// 保留该索引之后的所有元素
	relevantPrices := v.Prices[index:]

	logger.Debug("Prices length %d, relevantPrices length", len(v.Prices), len(relevantPrices))
	if len(relevantPrices) == 0 {
		return 0.0
	}

	minPrice, maxPrice := relevantPrices[0].Value, relevantPrices[0].Value
	for _, price := range relevantPrices {
		if price.Value < minPrice {
			minPrice = price.Value
		}
		if price.Value > maxPrice {
			maxPrice = price.Value
		}
	}

	return (maxPrice - minPrice) / minPrice
}

func (v *Volatility) ClearExpiredPrices(keepPricesMs int64) {
	if len(v.Prices) > 0 {
		cutoffTime := time.Now().Add(-time.Duration(keepPricesMs) * time.Millisecond)

		// 使用sort.Search找到第一个不满足要求的时间点的索引
		index := sort.Search(len(v.Prices), func(i int) bool {
			return !v.Prices[i].UpdateTime.Before(cutoffTime)
		})

		// 保留该索引之后的所有元素
		v.Prices = v.Prices[index:]
	}
}

type SymbolVolatilities struct {
	mu              sync.RWMutex
	KeepPricesMs    int64
	CalVolatilityMs int64
	Volatilities    map[string]*Volatility
}

func NewSymbolVolatilities(keepPricesMs int64, calVolatilityMs int64) *SymbolVolatilities {
	return &SymbolVolatilities{
		KeepPricesMs:    keepPricesMs,
		CalVolatilityMs: calVolatilityMs,
		Volatilities:    make(map[string]*Volatility),
	}
}

func (sv *SymbolVolatilities) GetVolatility(symbol string) (*Volatility, bool) {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	volatility, ok := sv.Volatilities[symbol]
	if !ok {
		return nil, false
	}
	return volatility, true
}

func (sv *SymbolVolatilities) UpdateVolatility(symbol string, price model.Price) {
	sv.mu.Lock()
	defer sv.mu.Unlock()
	volatility, ok := sv.Volatilities[symbol]

	if !ok {
		volatility = NewVolatility()
		sv.Volatilities[symbol] = volatility
	}
	volatility.Prices = append(volatility.Prices, price)
	volatility.Value = volatility.CalculateVolatility(sv.CalVolatilityMs)
	volatility.ClearExpiredPrices(sv.KeepPricesMs)
	sv.Volatilities[symbol] = volatility

}

func ShowSymbolVolatilities() {
	symbolVolatilities.mu.RLock()
	defer symbolVolatilities.mu.RUnlock()

	for symbol, volatility := range symbolVolatilities.Volatilities {
		logger.Info("Symbol: %s, Volatility: %f, Price.length: %d", symbol, volatility.Value, len(volatility.Prices))
	}
}
