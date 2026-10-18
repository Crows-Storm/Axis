package bitget

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Crows-Storm/Axis/common/contextx"
	"github.com/gorilla/websocket"
)

const prodBaseURL = "https://api.bitget.com"

type Client struct {
	baseURL    string
	wsURL      string
	httpClient *http.Client
	mu         sync.Mutex
	conn       *websocket.Conn
	cancel     context.CancelFunc
}

func NewClient() *Client {
	return &Client{
		baseURL:    prodBaseURL,
		wsURL:      "wss://ws.bitget.com/v2/ws/public",
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

type envelope struct {
	Code string          `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func (c *Client) get(ctx context.Context, path string, query url.Values, signed bool, out any) error {
	return c.do(ctx, http.MethodGet, path, query, nil, signed, out)
}

func (c *Client) post(ctx context.Context, path string, body any, out any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return c.do(ctx, http.MethodPost, path, nil, raw, true, out)
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body []byte, signed bool, out any) error {
	reqPath := path
	if len(query) > 0 {
		reqPath = path + "?" + query.Encode()
	}
	u := strings.TrimRight(c.baseURL, "/") + reqPath

	var rdr io.Reader
	if len(body) > 0 {
		rdr = bytes.NewReader(body)
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, u, rdr)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("locale", "en-US")
	if len(body) > 0 {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	if signed {
		cred, err := signedCredentials(ctx)
		if err != nil {
			return err
		}
		ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
		httpReq.Header.Set("ACCESS-KEY", cred.APIKey)
		httpReq.Header.Set("ACCESS-PASSPHRASE", cred.Passphrase)
		httpReq.Header.Set("ACCESS-TIMESTAMP", ts)
		httpReq.Header.Set("ACCESS-SIGN", signBitget(cred.SecretKey, ts, method, reqPath, body))
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("bitget decode: %w", err)
	}
	if env.Code != "00000" {
		msg := strings.TrimSpace(env.Msg)
		if msg == "" {
			msg = string(raw)
		}
		return fmt.Errorf("bitget api code %s: %s", env.Code, msg)
	}
	if out == nil || len(env.Data) == 0 || string(env.Data) == "null" {
		return nil
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("bitget decode data: %w", err)
	}
	return nil
}

func signedCredentials(ctx context.Context) (*contextx.Credentials, error) {
	cred, ok := contextx.CredentialsFromContext(ctx)
	if !ok || cred.APIKey == "" || cred.SecretKey == "" || cred.Passphrase == "" {
		return nil, errors.New("missing api credentials in context")
	}
	return cred, nil
}

func signBitget(secret, timestamp, method, requestPath string, body []byte) string {
	var b strings.Builder
	b.WriteString(timestamp)
	b.WriteString(method)
	b.WriteString(requestPath)
	b.Write(body)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(b.String()))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
