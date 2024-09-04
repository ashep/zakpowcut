package parser

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
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

func FileChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func ParseImage(path string, l zerolog.Logger) (TimeTable, error) {
	var (
		src image.Image
		err error
	)

	res := TimeTable{}

	fp, err := os.Open(path)
	if err != nil {
		return res, fmt.Errorf("failed to open source file: %w", err)
	}

	defer func() {
		if err = fp.Close(); err != nil {
			l.Error().Err(err).Msg("failed to close source file")
		}
	}()

	if strings.HasSuffix(strings.ToLower(path), ".png") {
		src, err = png.Decode(fp)
		if err != nil {
			return res, fmt.Errorf("failed to decode image: %w", err)
		}
	} else if strings.HasSuffix(strings.ToLower(path), ".jpg") || strings.HasSuffix(strings.ToLower(path), ".jpeg") {
		src, err = jpeg.Decode(fp)
		if err != nil {
			return res, fmt.Errorf("failed to decode image: %w", err)
		}
	}

	if src.Bounds().Dy() > 400 {
		return res, fmt.Errorf("image height is too big: %s", src.Bounds())
	}

	gray := image.NewGray(src.Bounds())
	draw.Draw(gray, gray.Bounds(), src, src.Bounds().Min, draw.Src)
	if l.GetLevel() == zerolog.DebugLevel {
		gfp, err := os.Create(path + "-gray.png")
		if err != nil {
			return res, fmt.Errorf("failed to open gray file: %w", err)
		}
		defer func() {
			if err = gfp.Close(); err != nil {
				l.Error().Err(err).Msg("failed to close gray file")
			}
		}()
		if err = png.Encode(gfp, gray); err != nil {
			return res, fmt.Errorf("failed to encode gray image: %w", err)
		}
	}

	cropXStart, cropYStart := 0, 0
lp1:
	for y := 0; y < gray.Bounds().Max.Y; y++ {
		for x := 0; x < gray.Bounds().Max.X; x++ {
			r, _, _, _ := gray.At(x, y).RGBA()
			if r>>8 <= 0x36 {
				cropXStart, cropYStart = x, y
				break lp1
			}
		}
	}

	cropXEnd, cropYEnd := gray.Bounds().Max.X-1, gray.Bounds().Max.Y-1
lp2:
	for y := gray.Bounds().Max.Y - 1; y >= 0; y-- {
		for x := gray.Bounds().Max.X - 1; x >= 0; x-- {
			r, _, _, _ := gray.At(x, y).RGBA()
			if r>>8 <= 0x36 {
				cropXEnd, cropYEnd = x, y
				break lp2
			}
		}
	}

	cropRect := image.Rect(cropXStart, cropYStart, cropXEnd, cropYEnd)
	cropped := image.NewRGBA(image.Rect(0, 0, cropXEnd-cropYStart, cropYEnd-cropYStart))
	draw.Copy(cropped, image.Point{X: 0, Y: 0}, gray, cropRect, draw.Over, nil)
	l.Debug().Msg(fmt.Sprintf("image cropped: %s -> %s", gray.Bounds().String(), cropped.Bounds().String()))

	if l.GetLevel() == zerolog.DebugLevel {
		cfp, err := os.Create(path + "-crop.png")
		if err != nil {
			return res, fmt.Errorf("failed to open cropped file: %w", err)
		}
		defer func() {
			if err = cfp.Close(); err != nil {
				l.Error().Err(err).Msg("failed to close cropped file")
			}
		}()
		if err = png.Encode(cfp, cropped); err != nil {
			return res, fmt.Errorf("failed to encode cropped image: %w", err)
		}
	}

	if src.Bounds().Dx()-cropRect.Bounds().Dx() > 50 {
		return res, fmt.Errorf("crop width is too big: %s -> %s", src.Bounds().String(), cropRect.Bounds().String())
	}
	if src.Bounds().Dy()-cropRect.Bounds().Dy() > 50 {
		return res, fmt.Errorf("crop height is too big: %s -> %s", src.Bounds().String(), cropRect.Bounds().String())
	}

	scaled := image.NewRGBA(image.Rect(0, 0, 1382, 305))
	draw.NearestNeighbor.Scale(scaled, scaled.Rect, cropped, cropped.Bounds(), draw.Over, nil)
	l.Info().Msg(fmt.Sprintf("image scaled: %s -> %s", cropped.Bounds().String(), scaled.Rect.String()))

	if l.GetLevel() == zerolog.DebugLevel {
		sfp, err := os.Create(path + "-scale.png")
		if err != nil {
			return res, fmt.Errorf("failed to open scaled file: %w", err)
		}
		defer func() {
			if err = sfp.Close(); err != nil {
				l.Error().Err(err).Msg("failed to close scaled file")
			}
		}()
		if err = png.Encode(sfp, scaled); err != nil {
			return res, fmt.Errorf("failed to encode scaled image: %w", err)
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
			// case isGray(r, g, b):
			// 	res[qn][hn] = PowerPerhaps
			default:
				res[qn][hn] = PowerOff
			}

			l.Debug().Msg(fmt.Sprintf("color: x=%d, y=%d, queue=%d, hour=%d, r=%d, g=%d, b=%d", x, y, qn+1, hn, r, g, b))
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
