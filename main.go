package main

import (
	"fmt"
	"net/http"
	"okx_data/common"
	"okx_data/common/logger"
	"okx_data/config"
	"os"
	"time"
)

var cfg config.Config
var eventHandler EventHandler
var symbolVolatilities *SymbolVolatilities

func Init(conf *config.Config) {
	eventHandler.Init(conf)
	symbolVolatilities = NewSymbolVolatilities(conf.KeepPricesMs, conf.CalVolatilityMs)
}

func Start() {
	// 启动websockets
	eventHandler.Start()

	time.Sleep(100 * time.Second)

	go common.Timer(60*time.Second, ShowSymbolVolatilities)
}

func ExitProcess() {
	// 取消所有订单, 不判断本地orders
	logger.Info("Disconnect all websocket.")
	eventHandler.Stop()
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s config_file\n", os.Args[0])
		os.Exit(1)
	}

	common.RegisterExitSignal(ExitProcess)

	// 加载配置文件
	cfg = *config.LoadConfig(os.Args[1])

	// 设置日志级别, 并初始化日志
	logger.InitLogger(cfg.LogPath, cfg.LogLevel)

	Init(&cfg)

	Start()

	http.HandleFunc("/api/v1/getSymbolVolatility", getSymbolVolatilityHandler())

	logger.Info("Server listening on :8888")
	http.ListenAndServe(":8888", nil)
}
