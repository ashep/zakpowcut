package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ashep/aghpu/httpclient"
	"github.com/ashep/aghpu/logger"
	"zakpowcut/printer"
	"zakpowcut/state"
	"zakpowcut/tg"

	"zakpowcut/dl"
	"zakpowcut/parser"
)

var qTgChannels = [4]string{
	"@uzhenergy1",
	"@uzhenergy2",
	"@uzhenergy3",
	"@uzhenergy4",
}

func main() {
	dbg := flag.Bool("d", false, "debug")
	flag.Parse()

	ll := logger.LvInfo
	if *dbg {
		ll = logger.LvDebug
	}

	l, err := logger.New(time.Now().Format("2006-01-02"), ll, "./log", "")
	if err != nil {
		panic(err)
	}
	defer l.Info("finished\n")

	st, err := state.Get()
	if err != nil {
		l.Warn("failed to read state: %s", err)
		if err = state.Save(st); err != nil {
			l.Err("failed to create a new state: %s", err)
			return
		}
		l.Info("a new state created")
	}
	l.Info("last file date: %s", st.String())

	httpCli, err := httpclient.New("zakpowcut", "./tmp", "", "", true, l)
	if err != nil {
		l.Err("failed to initialize an http client: %s", err)
		return
	}
	httpCli.SetMaxRetries(1)

	tgCli := tg.NewClient(httpCli, os.Getenv("TG_TOKEN"), l)
	if err = tgCli.Ping(context.Background()); err != nil {
		l.Err("failed to connect to telegram: %s", err)
		return
	}

	fPath, err := dl.GetImage(context.Background(), httpCli, l)
	if err != nil {
		l.Err("failed to download image: %s", err)
		return
	}
	l.Info("image downloaded: %s", fPath)

	dt, err := parser.ParseFileDate(fPath)
	if err != nil {
		l.Err("failed to determine file date: %s", err)
		return
	}
	l.Info("actual file date: %s", dt.String())

	if dt.Sub(st).String() == "0s" {
		l.Info("this file has already been processed")
		return
	}

	tt, err := parser.ParseImage(fPath, l)
	if err != nil {
		l.Err("failed to parse image: %s", err)
		return
	}
	l.Info("time table:\n%s", printer.PrintTimeTable(tt))

	trs := parser.TimeTableToTimeRanges(tt)
	l.Info("time ranges:\n%s", printer.PrintAllTimeRanges(trs, dt, "Черга", "`"))

	for qn, tr := range trs {
		msg := fmt.Sprintf("*Графік на %s*\n\n%s", dt.Format("02\\.01\\.2006"), printer.PrintTimeRanges(tr, "`"))
		if err = tgCli.SendMessage(context.Background(), qTgChannels[qn], msg); err != nil {
			l.Err("failed to send a message to to telegram: %s", err)
			return
		}
	}

	if err = state.Save(dt); err != nil {
		l.Err("failed to save state: %s", err)
		return
	}
	l.Info("state updated: %s", dt.String())
}
