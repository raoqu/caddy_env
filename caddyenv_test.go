package caddyenv

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/zeebo/assert"
)

func TestUnmarshalCaddyfile(t *testing.T) {
	// 定义 Caddyfile 的模拟配置
	caddyfileContent := `
    queryenv {
        url http://127.0.0.1:7788/wx_notify
    }
    `

	// 创建 Caddyfile 解析器
	d := caddyfile.NewTestDispenser(caddyfileContent)

	// 创建 CaddyEnv 实例
	var env CaddyEnv

	// 调用 UnmarshalCaddyfile 方法
	err := env.UnmarshalCaddyfile(d)
	assert.NoError(t, err)

	// 验证解析结果
	assert.Equal(t, "http://127.0.0.1:7788/wx_notify", env.URL)
}

func TestBodyProcessor_ServeHTTP(t *testing.T) {
	m := CaddyEnv{
		URL: "http://localhost:7788/wx_notify",
	}

	reqBody := `<xml><out_trade_no><![CDATA[GM240824133452586424]]></out_trade_no></xml>`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	next := caddyhttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		uri := r.Context().Value("uri")
		service := r.Context().Value("service")

		assert.Equal(t, "/v2/weChatNotify", uri)
		assert.Equal(t, "http://10.6.0.14:18099", service)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
		return nil
	})

	// 调用 ServeHTTP 方法
	err := m.ServeHTTP(rr, req, next)
	assert.NoError(t, err)

	// 检查响应码
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "success", rr.Body.String())
}
