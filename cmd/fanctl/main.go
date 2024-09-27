package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	"github.com/IvanSafonov/fanctl/internal/config"
	"github.com/IvanSafonov/fanctl/internal/service"
)

func main() {
	timeInLogs := false
	logLevel := new(slog.LevelVar)
	logLevel.Set(slog.LevelInfo)
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if !timeInLogs && a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}))
	slog.SetDefault(log)

	var (
		confPath   string
		debug      bool
		cpuprofile string
	)

	flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")
	flag.StringVar(&confPath, "c", "/etc/fanctl.yaml", "configuraion file path")
	flag.BoolVar(&debug, "d", false, "print debug messages")
	flag.Parse()

	if debug {
		logLevel.Set(slog.LevelDebug)
		timeInLogs = true
	}

	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Error("failed to create cpuprofile file", "err", err)
		} else {
			if err := pprof.StartCPUProfile(f); err != nil {
				log.Error("failed to start cpu profile", "err", err)
			}
			defer pprof.StopCPUProfile()
		}
	}

	conf, err := config.Load(confPath)
	if err != nil {
		slog.Error("config load error", "err", err)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := service.New(conf)

	err = srv.Init()
	if err != nil {
		slog.Error("service init error", "err", err)
		return
	}

	err = srv.Run(ctx)
	if err != nil {
		slog.Error("service run error", "err", err)
		return
	}
}
