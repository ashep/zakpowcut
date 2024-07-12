package parser

import (
	"fmt"
	"image"
	"image/png"
	"math"
	"os"

	"github.com/ashep/aghpu/logger"
	"golang.org/x/image/draw"
)

const (
	startX = 180
	startY = 85
	stepX  = 51
	stepY  = 40
)

const (
	PowerOff PowerState = iota
	PowerOn
	PowerPerhaps
)

type PowerState int

type TimeTable [6][24]PowerState

type TimeRange struct {
	State PowerState
	Start int
	End   int
}

type TimeRanges []TimeRange

func ParseImage(path string, l *logger.Logger) (TimeTable, error) {
	res := TimeTable{}

	fp, err := os.Open(path)
	if err != nil {
		return res, fmt.Errorf("failed to open source file: %w", err)
	}

	src, err := png.Decode(fp)
	if err != nil {
		return res, fmt.Errorf("failed to decode image: %w", err)
	}

	if err = fp.Close(); err != nil {
		return res, fmt.Errorf("failed to close source file: %w", err)
	}

	if src.Bounds().Dy() > 400 {
		return res, fmt.Errorf("image height is too big: %s", src.Bounds())
	}

	cropXStart, cropYStart := 0, 0
lp1:
	for y := 0; y < src.Bounds().Max.Y; y++ {
		for x := 0; x < src.Bounds().Max.X; x++ {
			r, g, b, a := src.At(x, y).RGBA()
			if r == 0 && g == 0 && b == 0 && a == 65535 {
				cropXStart, cropYStart = x, y
				break lp1
			}
		}
	}

	cropXEnd, cropYEnd := src.Bounds().Max.X, src.Bounds().Max.Y
lp2:
	for y := src.Bounds().Max.Y; y >= 0; y-- {
		for x := src.Bounds().Max.X; x >= 0; x-- {
			r, g, b, a := src.At(x, y).RGBA()
			if r == 0 && g == 0 && b == 0 && a == 65535 {
				cropXEnd, cropYEnd = x, y
				break lp2
			}
		}
	}

	cropRect := image.Rect(cropXStart, cropYStart, cropXEnd, cropYEnd)
	if src.Bounds().Dx()-cropRect.Bounds().Dx() > 50 {
		return res, fmt.Errorf("crop width is too big: %s -> %s", src.Bounds().String(), cropRect.Bounds().String())
	}
	if src.Bounds().Dy()-cropRect.Bounds().Dy() > 50 {
		return res, fmt.Errorf("crop height is too big: %s -> %s", src.Bounds().String(), cropRect.Bounds().String())
	}

	cropped := image.NewRGBA(image.Rect(0, 0, cropXEnd-cropYStart, cropYEnd-cropYStart))
	draw.Copy(cropped, image.Point{X: 0, Y: 0}, src, cropRect, draw.Over, nil)
	l.Debug("image cropped: %s -> %s", src.Bounds().String(), cropped.Bounds().String())

	if l.Level() == logger.LvDebug {
		if fp, err = os.Create(path + "-crop.png"); err != nil {
			return res, fmt.Errorf("failed to open cropped file: %w", err)
		}
		if err = png.Encode(fp, cropped); err != nil {
			return res, fmt.Errorf("failed to encode cropped image: %w", err)
		}
		if err = fp.Close(); err != nil {
			return res, fmt.Errorf("failed to close cropped file: %w", err)
		}
	}

	scaled := image.NewRGBA(image.Rect(0, 0, 1382, 305))
	draw.NearestNeighbor.Scale(scaled, scaled.Rect, cropped, cropped.Bounds(), draw.Over, nil)
	l.Info("image scaled: %s -> %s", cropped.Bounds().String(), scaled.Rect.String())

	if l.Level() == logger.LvDebug {
		if fp, err = os.Create(path + "-scale.png"); err != nil {
			return res, fmt.Errorf("failed to open scaled file: %w", err)
		}
		if err = png.Encode(fp, scaled); err != nil {
			return res, fmt.Errorf("failed to encode scaled image: %w", err)
		}
		if err = fp.Close(); err != nil {
			return res, fmt.Errorf("failed to close scaled file: %w", err)
		}
	}

	for qn := 0; qn < len(res); qn++ {
		for hn := 0; hn < 24; hn++ {
			x := startX + stepX*hn
			y := startY + stepY*qn
			r, g, b, _ := scaled.At(x, y).RGBA()
			r, g, b = r/256, g/256, b/256

			switch {
			case r == 255 && g == 255 && b == 255:
				res[qn][hn] = PowerOn
			case isGray(r, g, b):
				res[qn][hn] = PowerPerhaps
			default:
				res[qn][hn] = PowerOff
			}

			l.Debug("color: x=%d, y=%d, queue=%d, hour=%d, r=%d, g=%d, b=%d", x, y, qn+1, hn, r, g, b)
		}
	}

	return res, nil
}

func TimeTableToTimeRanges(tt TimeTable) []TimeRanges {
	var r = make([]TimeRanges, len(tt))

	for qn := 0; qn < len(tt); qn++ {
		r[qn] = make([]TimeRange, 0)

		cur := TimeRange{State: tt[qn][0]}
		for hn := 1; hn < 24; hn++ {
			if tt[qn][hn] != cur.State {
				cur.End = hn - 1
				r[qn] = append(r[qn], cur)
				cur = TimeRange{State: tt[qn][hn], Start: hn}
			} else {
				cur.End = hn
			}
		}

		cur.End = 23
		r[qn] = append(r[qn], cur)
	}

	return r
}

func isGray(r, g, b uint32) bool {
	rg := math.Abs(float64(int(r) - int(g)))
	rb := math.Abs(float64(int(r) - int(b)))
	gb := math.Abs(float64(int(g) - int(b)))

	return (rg+rb+gb)/3 < 7.0
}
