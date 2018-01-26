package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/fatih/color"

	"github.com/alfred-zhong/goscatter"
	"github.com/google/gops/agent"
	"github.com/jinzhu/configor"
)

type config struct {
	Port       int
	RemoteAddr string
	Scatters   []string
}

const usage = `usage: tcp-scatter [--gops-port=?] conf.json`

func main() {
	gopsPort := flag.Int("gops-port", 0, "port used to listen by gops")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println(usage)
		os.Exit(1)
	}

	// gops
	if *gopsPort > 0 {
		gopsOptions := agent.Options{
			Addr:            fmt.Sprintf("127.0.0.1:%d", *gopsPort),
			ShutdownCleanup: true,
		}
		if err := agent.Listen(gopsOptions); err != nil {
			fmt.Fprintf(os.Stderr, "gops listen on %d fail\n", *gopsPort)
			os.Exit(5)
		}
		color.Magenta("gops listen on port: %d\n", *gopsPort)
	}

	// load config file
	var cfg config
	if err := configor.Load(&cfg, flag.Arg(0)); err != nil {
		fmt.Fprintf(os.Stderr, "load config from %s fail\n", flag.Arg(0))
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
