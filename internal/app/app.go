package app

import (
	"context"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/ashep/go-httpcli"
	"github.com/rs/zerolog"

	"zakpowcut/internal/dl"
	"zakpowcut/internal/parser"
	"zakpowcut/internal/printer"
	"zakpowcut/internal/state"
	"zakpowcut/internal/tg"
)

var qTgChannels = []string{
	"@uzhenergy1",
	"@uzhenergy2",
	"@uzhenergy3",
	"@uzhenergy4",
	"@uzhenergy5",
	"@uzhenergy6",
}

type App struct {
	cfg Config
	l   zerolog.Logger
}

func New(cfg Config, l zerolog.Logger) *App {
	return &App{
		cfg: cfg,
		l:   l,
	}
}

func (a *App) Run(ctx context.Context, _ []string) error {
	if err := a.run(ctx); err != nil {
		a.l.Error().Err(err).Msg("run failed")
	}

	if a.cfg.Once {
		return nil
	}

	t := time.NewTicker(time.Hour)
	for {
		a.l.Info().Time("next_run_at", time.Now().Add(time.Hour)).Msg("sleeping")

		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			if err := a.run(ctx); err != nil {
				a.l.Error().Err(err).Msg("run failed")
			}
		}
	}
}

func (a *App) run(ctx context.Context) error {
	httpCli := httpcli.New(a.l)

	if len(a.cfg.ProxyURLs) != 0 {
		httpCli.SetProxyURLs(a.cfg.ProxyURLs)
		httpCli.SetMaxTries(len(a.cfg.ProxyURLs))
	}

	for _, d := range []string{"log", "tmp"} {
		if err := os.MkdirAll(d, 0755); err != nil {
			panic(err)
		}
	}

	now := time.Now()

	stateData, err := state.Get()
	if err != nil {
		a.l.Warn().Err(err).Msg("failed to read state")
		if err = state.Save(map[string][]string{}); err != nil {
			return fmt.Errorf("failed to create a new state: %w", err)
		}
		stateData = map[string][]string{}
	}

	fPaths, err := dl.GetImages(context.Background(), httpCli, a.l)
	if err != nil {
		a.l.Error().Err(err).Msg("")
		return fmt.Errorf("failed to get images: %w", err)
	}
	a.l.Info().Strs("images", fPaths).Msg("images downloaded")

	imgDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	for i, fPath := range fPaths {
		if i == 0 && a.cfg.SkipToday {
			a.l.Info().Msg("skip today")
			continue
		}

		imgDate = imgDate.Add(time.Hour * 24 * time.Duration(i)) // first image is for today, second image is for tomorrow
		imgDateStr := imgDate.Format("2006-01-02")

		imgHash, err := parser.FileChecksum(fPath)
		if err != nil {
			a.l.Error().Err(err).Str("path", fPath).Msg("failed to get image hash")
			continue
		}

		imgHashes, todayExists := stateData[imgDateStr]
		isUpdate := false
		if todayExists {
			if slices.Contains(imgHashes, imgHash) {
				a.l.Info().Str("date", imgDateStr).Str("hash", imgHash).Msg("image has already been processed")
				continue
			}
			isUpdate = true
		} else {
			imgHashes = make([]string, 0)
		}

		tt, err := parser.ParseImage(fPath, a.l)
		if err != nil {
			return fmt.Errorf("failed to parse image: %w", err)
		}
		a.l.Info().Msgf("time table:\n%s", printer.PrintTimeTable(tt))

		title := fmt.Sprintf("*Графік на %s", imgDate.Format("02\\.01\\.2006"))
		if isUpdate {
			title += " (оновлено)"
		}
		title += "*"

		if a.cfg.DryRun {
			for qn, tr := range parser.TimeTableToTimeRanges(tt) {
				msg := fmt.Sprintf("Черга %d\n", qn+1)
				msg += fmt.Sprintf("%s\n\n%s", title, printer.PrintTimeRanges(tr))
				a.l.Debug().Msgf("\n%s", msg)
			}
		} else {
			tgCli := tg.NewClient(httpCli, a.cfg.TgToken)
			if err = tgCli.Ping(ctx); err != nil {
				return fmt.Errorf("failed to connect to telegram: %w", err)
			}

			for qn, tr := range parser.TimeTableToTimeRanges(tt) {
				msg := fmt.Sprintf("%s\n\n%s", title, printer.PrintTimeRanges(tr))
				if err = tgCli.SendMessage(context.Background(), qTgChannels[qn], msg); err != nil {
					return fmt.Errorf("failed to send a message to telegram: %w", err)
				}
			}
		}

		imgHashes = append(imgHashes, imgHash)

		stateData[imgDateStr] = imgHashes
		if err = state.Save(stateData); err != nil {
			return fmt.Errorf("failed to save state: %w", err)
		}
	}

	return nil
}
