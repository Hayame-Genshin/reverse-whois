package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"

	"github.com/haltman-io/reverse-whois/internal/util"
)

type ClientOptions struct {
	Endpoint  string
	ProxyURL  string
	NoProxy   bool
	Insecure  bool
	Timeout   time.Duration
	UserAgent string
	Logger    *util.Logger
}

type Client struct {
	endpoint   string
	httpClient *http.Client
	userAgent  string
	logger     *util.Logger
}

func NewClient(opts ClientOptions) (*Client, error) {
	if opts.Endpoint == "" {
		return nil, fmt.Errorf("api endpoint is required")
	}

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   15 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
	}

	// Proxy behavior: --no-proxy overrides everything.
	if opts.NoProxy {
		tr.Proxy = nil
	} else if opts.ProxyURL != "" {
		u, err := url.Parse(opts.ProxyURL)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return nil, fmt.Errorf("invalid --proxy URL: %q", opts.ProxyURL)
		}

		switch u.Scheme {
		case "http", "https":
			tr.Proxy = http.ProxyURL(u)
		case "socks5":
			var auth *proxy.Auth
			if u.User != nil {
				pass, _ := u.User.Password()
				auth = &proxy.Auth{
					User:     u.User.Username(),
					Password: pass,
				}
			}
			d, err := proxy.SOCKS5("tcp", u.Host, auth, proxy.Direct)
			if err != nil {
				return nil, fmt.Errorf("failed to configure socks5 proxy: %w", err)
			}

			tr.Proxy = nil
			tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
				return d.Dial(network, addr)
			}
		default:
			return nil, fmt.Errorf("unsupported proxy scheme %q (use http, https, or socks5)", u.Scheme)
		}
	}

	// TLS insecure (curl -k)
	if opts.Insecure {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	hc := &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}

	return &Client{
		endpoint:   opts.Endpoint,
		httpClient: hc,
		userAgent:  opts.UserAgent,
		logger:     opts.Logger,
	}, nil
}

func (c *Client) Search(ctx context.Context, req Request) (Response, int, error) {
	var out Response

	body, err := json.Marshal(req)
	if err != nil {
		return out, 0, fmt.Errorf("failed to marshal request json: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return out, 0, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.userAgent != "" {
		httpReq.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return out, 0, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return out, resp.StatusCode, &HTTPError{StatusCode: resp.StatusCode}
	}

	// Permissive decoding by default (provider may add fields over time).
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&out); err != nil {
		return Response{}, resp.StatusCode, fmt.Errorf("failed to decode response json: %w", err)
	}

	return out, resp.StatusCode, nil
}
