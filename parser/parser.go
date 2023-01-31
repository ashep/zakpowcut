package parser

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path"
	"regexp"
	"strconv"
	"time"

	"github.com/ashep/aghpu/logger"
	"golang.org/x/image/draw"
)

const (
	startX = 184
	startY = 122
	stepX  = 43
	stepY  = 42
)

const (
	PowerOff PowerState = iota
	PowerOn
	PowerPerhaps
)

type PowerState int

type TimeTable [4][24]PowerState

type TimeRange struct {
	State PowerState
	Start int
	End   int
}

type TimeRanges []TimeRange

func ParseImage(path string, l *logger.Logger) (TimeTable, error) {
	r := TimeTable{}

	fp, err := os.Open(path)
	if err != nil {
		return r, fmt.Errorf("failed to open file: %w", err)
	}

	src, err := png.Decode(fp)
	if err != nil {
		return r, fmt.Errorf("failed to decode image: %w", err)
	}

	if err = fp.Close(); err != nil {
		return r, fmt.Errorf("failed to close file: %w", err)
	}

	if src.Bounds().Dy() > 330 {
		return r, fmt.Errorf("image height is too big: %s", src.Bounds())
	}

	dst := image.NewRGBA(image.Rect(0, 0, 1200, 272))
	draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

	l.Info("image scaled: %s -> %s", src.Bounds().String(), dst.Rect.String())

	for qn := 0; qn < 4; qn++ {
		for hn := 0; hn < 24; hn++ {
			x := startX + stepX*hn
			y := startY + stepY*qn
			cr, cg, cb, _ := dst.At(x, y).RGBA()

			l.Debug("color: queue=%d, hour=%d, r=%d, g=%d, b=%d", qn+1, hn, cr, cg, cb)

			switch {
			case cr == 65535 && cg == 65535 && cb == 65535:
				r[qn][hn] = PowerOn
			case cr == 32896 && cg == 32896 && cb == 32896:
				r[qn][hn] = PowerPerhaps
			default:
				r[qn][hn] = PowerOff
			}
		}
	}

	return r, nil
}

func ParseFileDate(pth string) (time.Time, error) {
	var r time.Time

	re, err := regexp.Compile(`\d+`)
	if err != nil {
		return r, err
	}

	_, n := path.Split(pth)
	rm := re.FindAllString(n, 1)
	if len(rm) == 0 {
		return r, fmt.Errorf("unexpected filename: %s", n)
	}

	if len(rm[0]) != 6 && len(rm[0]) != 8 {
		return r, fmt.Errorf("unexpected filename: %s", n)
	}

	day := rm[0][0:2]
	month := rm[0][2:4]
	year := "20" + rm[0][4:6]
	if len(rm[0]) == 8 {
		year = rm[0][4:8]
	}

	return time.Date(mustAtoi(year), time.Month(mustAtoi(month)), mustAtoi(day), 0, 0, 0, 0, time.UTC), nil
}

func TimeTableToTimeRanges(tt TimeTable) [4]TimeRanges {
	var r [4]TimeRanges

	for qn := 0; qn < 4; qn++ {
		r[qn] = make([]TimeRange, 0)

		cur := TimeRange{State: tt[qn][0]}
		for hn := 1; hn < 24; hn++ {
			if tt[qn][hn] != cur.State {
				cur.End = hn
				r[qn] = append(r[qn], cur)
				cur = TimeRange{State: tt[qn][hn], Start: hn}
			} else {
				cur.End = hn + 1
			}
		}

		cur.End = 24
		r[qn] = append(r[qn], cur)
	}

	return r
}

func mustAtoi(s string) int {
	r, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}

	return r
}
