package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/alfred-zhong/goscatter"
	"github.com/jinzhu/configor"
)

type config struct {
	Port       int
	RemoteAddr string
	Scatters   []string
}

const usage = `usage: tcp-scatter conf.json`

func main() {
	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(1)
	}

	// load config file
	var cfg config
	if err := configor.Load(&cfg, os.Args[1]); err != nil {
		fmt.Fprintf(os.Stderr, "load config from %s fail\n", os.Args[1])
		os.Exit(2)
	}

	// fmt.Println(config)
	server, err := goscatter.NewServer(cfg.Port, cfg.RemoteAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewServer fail: %v\n", err)
		os.Exit(3)
	}
	for _, scatter := range cfg.Scatters {
		server.AddScatterAddr(scatter)
	}

	// receive signal to stop the whole process
	killCh := make(chan os.Signal)
	signal.Notify(killCh, os.Interrupt, os.Kill)
	go func() {
		<-killCh
		server.Stop()
	}()

	// Run server
	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Run server fail: %v\n", err)
		os.Exit(4)
	}
}
