package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ashep/aghpu/httpclient"
	"github.com/ashep/aghpu/logger"
	"zakpowcut/dl"
	"zakpowcut/parser"
	"zakpowcut/printer"
	"zakpowcut/state"
	"zakpowcut/tg"
)

var qTgChannels = [4]string{
	"@uzhenergy1",
	"@uzhenergy2",
	"@uzhenergy3",
	"@uzhenergy4",
}

var buildName, buildVer string

func main() {
	dbg := flag.Bool("d", false, "debug mode")
	dry := flag.Bool("dry-run", false, "skip publishing step")
	flag.Parse()

	ll := logger.LvInfo
	if *dbg {
		ll = logger.LvDebug
	}

	for _, d := range []string{"log", "tmp"} {
		if err := os.MkdirAll(d, 0755); err != nil {
			panic(err)
		}
	}

	now := time.Now()

	l, err := logger.New(buildName+"-"+buildVer, ll, "./log", now.Format("2006-01-02")+".log")
	if err != nil {
		panic(err)
	}
	defer l.Info("finished\n")

	lastFPath, err := state.Get()
	if err != nil {
		l.Warn("failed to read state: %s", err)
		if err = state.Save(lastFPath); err != nil {
			l.Err("failed to create a new state: %s", err)
			return
		}
		l.Info("a new state has been created")
	}
	l.Info("last file: %s", lastFPath)

	httpCli, err := httpclient.New("zakpowcut", "./tmp", "", "", *dbg, l)
	if err != nil {
		l.Err("failed to initialize an http client: %s", err)
		return
	}
	httpCli.SetMaxRetries(1)

	fPath, err := dl.GetImage(context.Background(), httpCli, l)
	if err != nil {
		l.Err("failed to download image: %s", err)
		return
	}
	l.Info("image downloaded: %s", fPath)

	if lastFPath == fPath {
		l.Info("this file has already been processed")
		return
	}
	
	tt, err := parser.ParseImage(fPath, l)
	if err != nil {
		l.Err("failed to parse image: %s", err)
		return
	}
	l.Info("time table:\n%s", printer.PrintTimeTable(tt))

	if *dry {
		for qn, tr := range parser.TimeTableToTimeRanges(tt) {
			msg := fmt.Sprintf("Черга %d\n", qn+1)
			msg += fmt.Sprintf("*Графік на %s*\n\n%s", now.Format("02.01.2006"), printer.PrintTimeRanges(tr))
			l.Debug("\n%s", msg)
		}
		return
	}

	tgCli := tg.NewClient(httpCli, os.Getenv("TG_TOKEN"), l)
	if err = tgCli.Ping(context.Background()); err != nil {
		l.Err("failed to connect to telegram: %s", err)
		return
	}

	for qn, tr := range parser.TimeTableToTimeRanges(tt) {
		msg := fmt.Sprintf("*Графік на %s*\n\n%s", now.Format("02\\.01\\.2006"), printer.PrintTimeRanges(tr))
		if err = tgCli.SendMessage(context.Background(), qTgChannels[qn], msg); err != nil {
			l.Err("failed to send a message to telegram: %s", err)
			return
		}
	}

	if err = state.Save(fPath); err != nil {
		l.Err("failed to save state: %s", err)
		return
	}
	l.Info("state updated: %s", fPath)
}
