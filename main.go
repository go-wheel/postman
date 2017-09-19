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
)

func main() {

	// subscribe to SIGINT signals
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	logger.Init(nil)

	flag.IntVar(&proxyType, "proxy-type", 0, "proxy-type")
	flag.StringVar(&proxyURL, "proxy-url", "", "proxy-url")

	flag.Parse()

	httpclient.Settings().SetProxy(httpclient.ProxyType(proxyType), proxyURL)

	backendhttp.Init()

	<-stopChan // wait for SIGINT

	logger.Infoln("gracefully stopped stopped")

}
