package dl

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ashep/aghpu/httpclient"
	"github.com/ashep/aghpu/logger"
)

const (
	baseURL       = "https://zakarpat.energy"
	imagePagePath = "/customers/break-in-electricity-supply/schedule/"
)

func GetImage(ctx context.Context, cli *httpclient.Cli, l *logger.Logger) (string, error) {
	d, err := cli.GetQueryDoc(ctx, baseURL+imagePagePath, nil, nil)
	if err != nil {
		return "", fmt.Errorf("failed to download page: %w", err)
	}

	imgURL := ""
	d.Find(".news-single .container2 img").Each(func(i int, selection *goquery.Selection) {
		src, _ := selection.Attr("src")
		if strings.HasPrefix(src, "/upload/current-timetable/") && strings.HasSuffix(strings.ToLower(src), ".png") {
			if strings.HasPrefix(src, "http") {
				imgURL = src
			} else {
				imgURL = baseURL + src
			}
		}
	})

	if imgURL == "" {
		return "", errors.New("failed to find image on the page")
	}

	l.Debug("image found: %s", imgURL)

	imgURLSplit := strings.Split(imgURL, "/")
	fName := imgURLSplit[len(imgURLSplit)-1]
	fPath := "./tmp/" + fName

	if _, err = cli.GetFile(ctx, imgURL, nil, nil, fPath); err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}

	return fPath, nil
}
