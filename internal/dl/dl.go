package dl

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ashep/go-httpcli"
	"github.com/rs/zerolog"
)

const (
	baseURL       = "https://zakarpat.energy"
	imagePagePath = "/customers/break-in-electricity-supply/schedule/"
)

func GetImages(ctx context.Context, cli *httpcli.Client, l zerolog.Logger) ([]string, error) {
	d, err := cli.GetQueryDoc(ctx, baseURL+imagePagePath, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to download page: %w", err)
	}

	imgURLs := make([]string, 0)
	d.Find(".news-single .container2 img").Each(func(i int, selection *goquery.Selection) {
		src, _ := selection.Attr("src")

		if !strings.HasSuffix(strings.ToLower(src), ".png") {
			return
		}

		if strings.HasPrefix(src, "/upload/current-timetable/") || strings.HasPrefix(src, "/upload/timetable-now/") {
			u := src
			if !strings.HasPrefix(src, "http") {
				u = baseURL + u
			}
			imgURLs = append(imgURLs, u)
		}
	})

	if len(imgURLs) == 0 {
		return nil, errors.New("failed to find image on the page")
	}
	l.Info().Strs("urls", imgURLs).Msg("images found")

	fPaths := make([]string, 0)
	for _, imgURL := range imgURLs {
		imgURLSplit := strings.Split(imgURL, "/")
		fName := imgURLSplit[len(imgURLSplit)-1]
		fPath := "./tmp/" + fName

		if _, err = cli.GetFile(ctx, imgURL, nil, nil, fPath); err != nil {
			return nil, fmt.Errorf("failed to download image: %w", err)
		}

		fPaths = append(fPaths, fPath)
	}

	return fPaths, nil
}
