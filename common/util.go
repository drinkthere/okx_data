package common

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type SimpleFunc func()

func Timer(interval time.Duration, RunFunc SimpleFunc) {
	for {
		RunFunc()
		time.Sleep(interval)
	}
}

func RegisterExitSignal(exitFunc SimpleFunc) {
	c := make(chan os.Signal, 5)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				exitFunc()
			default:
				fmt.Println("其他信号:", s)
			}
		}
	}()
}
