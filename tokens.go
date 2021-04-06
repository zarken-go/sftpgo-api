package api

import (
	"time"

	"gopkg.in/h2non/gentleman.v2"
	"gopkg.in/h2non/gentleman.v2/context"
	"gopkg.in/h2non/gentleman.v2/plugin"
	"gopkg.in/h2non/gentleman.v2/plugins/auth"
	"gopkg.in/h2non/gentleman.v2/plugins/timeout"
)

type TokenProvider interface {
	Token() (string, error)
}

type tokenProvider struct {
	Value    string    `json:"access_token"`
	Expires  time.Time `json:"expires_at"`
	Client   *gentleman.Client
	Username string
	Password string
}

func (t *tokenProvider) Token() (string, error) {
	if t.Expires.After(time.Now().Add(time.Second * 10)) {
		return t.Value, nil
	}

	req := t.Client.Request()
	req.Path(`/api/v2/token`)
	req.Use(auth.Basic(t.Username, t.Password))
	resp, err := req.Do()
	if err != nil {
		return ``, err
	}
	if resp.Ok {
		if err := resp.JSON(t); err != nil {
			return ``, err
		}
		return t.Value, nil
	}

	return ``, parseErrorResponse(resp)
}

func NewTokenProvider(cli *gentleman.Client, Username, Password string) TokenProvider {
	return &tokenProvider{
		Client:   cli,
		Username: Username,
		Password: Password,
	}
}

func NewTokenPlugin(URL, Username, Password string) plugin.Plugin {
	cli := gentleman.New()
	cli.URL(URL)
	cli.Use(timeout.All(timeout.Timeouts{
		Request:   5 * time.Second,
		TLS:       5 * time.Second,
		Dial:      5 * time.Second,
		KeepAlive: 5 * time.Second,
	}))

	provider := NewTokenProvider(cli, Username, Password)

	return plugin.NewRequestPlugin(func(ctx *context.Context, h context.Handler) {
		if token, err := provider.Token(); err == nil {
			ctx.Request.Header.Set("Authorization", "Bearer "+token)
			h.Next(ctx)
		} else {
			h.Error(ctx, err)
		}
	})
}
