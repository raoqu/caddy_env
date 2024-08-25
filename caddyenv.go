package caddyenv

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

type CaddyEnv struct {
	URL string `json:"url,omitempty"`
}

func init() {
	caddy.RegisterModule(CaddyEnv{})
}

func (CaddyEnv) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.queryenv",
		New: func() caddy.Module { return new(CaddyEnv) },
	}
}

func (m CaddyEnv) Provision(ctx caddy.Context) error {
	if m.URL == "" {
		return errors.New("queryenv: url 参数是必需的")
	}
	return nil
}

func (m CaddyEnv) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return caddyhttp.Error(http.StatusBadRequest, err)
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body))

	context, err := m.processBody(r.Context(), body)
	r = r.WithContext(context)

	return next.ServeHTTP(w, r)
}

func (m *CaddyEnv) processBody(ctx context.Context, body []byte) (context.Context, error) {
	jsonMap, err := fetch(m.URL, body)
	if err != nil {
		return ctx, caddyhttp.Error(http.StatusBadGateway, err)
	}

	newContext := ctx
	for k, v := range jsonMap {
		newContext = context.WithValue(newContext, k, v)
	}
	return newContext, nil
}

func fetch(url string, body []byte) (map[string]string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second, // 设置超时时间
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	// req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("非200响应: " + res.Status)
	}

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]string
	err = json.Unmarshal(responseBody, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
