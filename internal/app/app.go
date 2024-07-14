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
	for _, d := range []string{"log", "tmp"} {
		if err := os.MkdirAll(d, 0755); err != nil {
			panic(err)
		}
	}

	now := time.Now()

	knownDates, err := state.Get()
	if err != nil {
		a.l.Warn().Err(err).Msg("failed to read state")
		if err = state.Save([]string{}); err != nil {
			return fmt.Errorf("failed to create a new state: %w", err)
		}
	}

	httpCli := httpcli.New(a.l)
	httpCli.SetMaxTries(1)

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

		// first image is for today, second image is fo tomorrow
		imgDate = imgDate.Add(time.Hour * 24 * time.Duration(i))

		imgDateStr := imgDate.Format("2006-01-02")
		if slices.Contains(knownDates, imgDateStr) {
			a.l.Info().Str("date", imgDateStr).Msg("date has already been processed")
			continue
		}

		tt, err := parser.ParseImage(fPath, a.l)
		if err != nil {
			return fmt.Errorf("failed to parse image: %w", err)
		}
		a.l.Info().Msgf("time table:\n%s", printer.PrintTimeTable(tt))

		if a.cfg.DryRun {
			for qn, tr := range parser.TimeTableToTimeRanges(tt) {
				msg := fmt.Sprintf("Черга %d\n", qn+1)
				msg += fmt.Sprintf("*Графік на %s*\n\n%s", imgDate.Format("02.01.2006"), printer.PrintTimeRanges(tr))
				a.l.Debug().Msgf("\n%s", msg)
			}
		} else {
			tgCli := tg.NewClient(httpCli, a.cfg.TgToken)
			if err = tgCli.Ping(ctx); err != nil {
				return fmt.Errorf("failed to connect to telegram: %w", err)
			}

			for qn, tr := range parser.TimeTableToTimeRanges(tt) {
				msg := fmt.Sprintf("*Графік на %s*\n\n%s", imgDate.Format("02\\.01\\.2006"), printer.PrintTimeRanges(tr))
				if err = tgCli.SendMessage(context.Background(), qTgChannels[qn], msg); err != nil {
					return fmt.Errorf("failed to send a message to telegram: %w", err)
				}
			}
		}

		knownDates = append(knownDates, imgDateStr)
		if err = state.Save(knownDates); err != nil {
			return fmt.Errorf("failed to save state: %w", err)
		}
	}

	return nil
}
