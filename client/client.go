package client

import "okx_data/model"

// 配置信息，用于Client初始化
type Config struct {
	Symbols []string // 多币种

	OkxAPIKey    string
	OkxSecretKey string
	OkxPassword  string
}

type OrderClient interface {
	Init(config Config) bool
}

// ------------------以下是websocket相关的内容

type PriceWSResponse struct {
	Symbol string
	Price  model.Price
}

// PriceProcessHandler 处理price ws消息返回的数据
type PriceProcessHandler func(resp *PriceWSResponse)

// ErrorHandler 处理ws消息抛出的错误
type ErrorHandler func(exchange string, error int)

type WSClient interface {
	Init(config Config) bool
	SetPriceHandler(handler PriceProcessHandler, errHandler ErrorHandler)
	StartWS() bool
	StopWS() bool
}
