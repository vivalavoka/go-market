package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/vivalavoka/go-market/cmd/gophermart/config"
	server "github.com/vivalavoka/go-market/cmd/gophermart/http"
	"github.com/vivalavoka/go-market/cmd/gophermart/storage"
)

func main() {
	var err error
	var cfg config.Config
	var stg *storage.Storage

	if cfg, err = config.Init(); err != nil {
		log.Fatal(err)
	}

	if stg, err = storage.New(cfg); err != nil {
		log.Fatal(err)
	}
	defer stg.Close()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		stg.Close()
		os.Exit(1)
	}()

	http := server.New(stg)
	http.Start(cfg)
}
