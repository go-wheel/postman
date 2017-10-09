package main

import (
	"flag"
	"os"
	"os/signal"

	"github.com/go-wheel/postman/backendhttp"

	"github.com/lets-go-go/httpclient"
	"github.com/lets-go-go/logger"
)

var (
	proxyType int
	proxyURL  string
	logFile   string
)

func main() {

	// subscribe to SIGINT signals
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	flag.IntVar(&proxyType, "proxy-type", 0, "proxy-type")
	flag.StringVar(&proxyURL, "proxy-url", "", "proxy-url")

	flag.StringVar(&logFile, "log-file", "./task.log", "log file")

	flag.Parse()

	config := logger.DefalutConfig()
	config.LogFileName = logFile

	logger.Init(config)

	httpclient.Settings().SetProxy(httpclient.ProxyType(proxyType), proxyURL)

	backendhttp.Init()

	<-stopChan // wait for SIGINT

	logger.Infoln("gracefully stopped stopped")

}
