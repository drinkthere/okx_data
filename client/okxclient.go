package client

import (
	"context"
	"github.com/drinkthere/okx"
	"github.com/drinkthere/okx/api"
	"github.com/drinkthere/okx/events"
	"github.com/drinkthere/okx/events/public"
	"github.com/drinkthere/okx/models/market"
	ws_public_requests "github.com/drinkthere/okx/requests/ws/public"
	"log"
	"okx_data/common/logger"
	"okx_data/model"
	"time"
)

type OkxClient struct {
	OrderClient
	WSClient

	priceWSHandler PriceProcessHandler
	errorHandler   ErrorHandler

	client  *api.Client
	symbols []string //多币种
}

func (cli *OkxClient) Init(config Config) bool {
	cli.symbols = config.Symbols

	dest := okx.NormalServer // The main API server
	ctx := context.Background()
	client, err := api.NewClient(ctx, config.OkxAPIKey, config.OkxSecretKey, config.OkxPassword, dest)
	if err != nil {
		log.Fatalln(err)
		return false
	}
	cli.client = client
	return true
}

func (cli *OkxClient) SetPriceHandler(handler PriceProcessHandler, errHandler ErrorHandler) {
	cli.priceWSHandler = handler
	cli.errorHandler = errHandler
}

func (cli *OkxClient) StartWS() bool {
	return cli.wSConnectNSubscribe()
}

func (cli *OkxClient) wSConnectNSubscribe() bool {
	errChan := make(chan *events.Error)
	subChan := make(chan *events.Subscribe)
	uSubChan := make(chan *events.Unsubscribe)
	lCh := make(chan *events.Login)
	succChan := make(chan *events.Success)

	tickerCh := make(chan *public.Tickers)
	cli.client.Ws.SetChannels(errChan, subChan, uSubChan, lCh, succChan)
	go func() {
		for {
			select {
			case sub := <-subChan:
				channel, _ := sub.Arg.Get("channel")
				logger.Info("[Subscribed]\t%s", channel)
			case err := <-errChan:
				logger.Warn("[Error]\t%+v", err)
				for _, datum := range err.Data {
					logger.Warn("[Error]\t\t%+v", datum)
				}
				time.Sleep(3 * time.Second) // 休眠1秒，重新连接
				cli.wSConnectNSubscribe()

			case i := <-tickerCh:
				cli.bookTickerMsgHandler(i.Tickers)

			case b := <-cli.client.Ws.DoneChan:
				logger.Info("[End]:\t%v", b)
			}
		}
	}()
	// 启动 WebSocket 订阅
	result := true
	for _, symbol := range cli.symbols {
		go func(sym string) {
			// 启动 futures 的 bookTicker
			err := cli.client.Ws.Public.Tickers(ws_public_requests.Tickers{
				InstID: sym,
			}, tickerCh)

			if err != nil {
				logger.Fatal(err.Error())
				result = false
				return
			}
			logger.Info("okx futures bookTicker WS is established, symbol:%s", sym)
		}(symbol)
	}
	return result
}

func (cli *OkxClient) StopWS() bool {
	cli.client.Ws.Cancel()
	return true
}

func (cli *OkxClient) bookTickerMsgHandler(tickers []*market.Ticker) {
	for _, ticker := range tickers {
		bidPrice := float64(ticker.BidPx)
		askPrice := float64(ticker.AskPx)
		if bidPrice > 0 && askPrice > 0 {
			var price model.Price
			// price.Value = (bidPrice + askPrice) / 2
			price.Value = bidPrice
			price.UpdateTime = time.Now()

			var priceResp PriceWSResponse
			priceResp.Symbol = ticker.InstID
			priceResp.Price = price

			logger.Debug("%s price:%.2f, time:%s", priceResp.Symbol, priceResp.Price.Value, priceResp.Price.UpdateTime)
			cli.priceWSHandler(&priceResp)
		}
	}
}
