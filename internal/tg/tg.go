package tg

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ashep/go-httpcli"
)

type Client struct {
	h *httpcli.Client
	t string
}

func NewClient(httpCli *httpcli.Client, token string) *Client {
	return &Client{
		h: httpCli,
		t: token,
	}
}

func (c *Client) Ping(ctx context.Context) error {
	if _, err := c.h.Get(ctx, c.u("getMe"), nil, nil); err != nil {
		return fmt.Errorf("ping failed: %s", err)
	}

	return nil
}

func (c *Client) SendMessage(ctx context.Context, chatID, text string) error {
	_, err := c.h.PostForm(ctx, c.u("sendMessage"), url.Values{
		"chat_id":    []string{chatID},
		"text":       []string{text},
		"parse_mode": []string{"MarkdownV2"},
	}, nil)

	return err
}

func (c *Client) u(method string) string {
	return fmt.Sprintf("https://api.telegram.org/bot%s/%s", c.t, method)
}
