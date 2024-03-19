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
	resultChan := make(chan bool)
	subChan := make(chan *events.Subscribe)
	errChan := make(chan *events.Error)
	tickerCh := make(chan *public.Tickers)

	go func() {
		for _, symbol := range cli.symbols {
			// 启动 futures的bookTicker
			err := cli.client.Ws.Public.Tickers(ws_public_requests.Tickers{
				InstID: symbol,
			}, tickerCh)

			if err != nil {
				logger.Fatal(err.Error())
			}
			logger.Info("okx futures bookTicker WS is established, symbol:%s", symbol)
			time.Sleep(50 * time.Millisecond)
		}

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
			case i := <-tickerCh:
				cli.bookTickerMsgHandler(i.Tickers)
			case b := <-cli.client.Ws.DoneChan:
				logger.Info("[End]:\t%v", b)
			}
		}
	}()

	go func() {
		resultChan <- true
	}()
	return <-resultChan
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
