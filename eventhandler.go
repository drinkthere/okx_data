package main

import (
	"okx_data/client"
	"okx_data/common"
	"okx_data/config"
)

type EventHandler struct {
	wsClient []client.WSClient
}

func (handler *EventHandler) Init(cfg *config.Config) {
	clientConfig := client.Config{
		OkxAPIKey:    cfg.OkxAPIKey,
		OkxSecretKey: cfg.OkxSecretKey,
		OkxPassword:  cfg.OkxPassword,
		Symbols:      cfg.Symbols,
	}

	// 初始化okx WS client
	okxWSClient := new(client.OkxClient)
	okxWSClient.Init(clientConfig)
	okxWSClient.SetPriceHandler(OkxPriceWSHandler, common.CommonErrorHandler)
	handler.wsClient = append(handler.wsClient, okxWSClient)
}

func (handler *EventHandler) Start() {
	for _, wsClient := range handler.wsClient {
		wsClient.StartWS()
	}
}

func (handler *EventHandler) Stop() {
	for _, wsClient := range handler.wsClient {
		wsClient.StopWS()
	}
}

func OkxPriceWSHandler(resp *client.PriceWSResponse) {
	// 更新SymbolVolatility
	symbolVolatilities.UpdateVolatility(resp.Symbol, resp.Price)
}
