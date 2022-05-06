package genesysgo

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
	"github.com/klauspost/compress/gzhttp"
)

var (
	TokenLifetime              = 5 * time.Minute
	defaultMaxIdleConnsPerHost = 20
	defaultTimeout             = 5 * time.Minute
	defaultKeepAlive           = 3 * time.Minute
)

type roundTripper struct {
	next http.RoundTripper

	mu           sync.Mutex
	refreshToken string
	accessToken  string
	fetchedAt    time.Time
}

func newRoundTripper(refreshToken string, next http.RoundTripper) *roundTripper {
	return &roundTripper{next: next, refreshToken: refreshToken}
}

func (rt *roundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if rt.refreshToken == "" {
		return rt.next.RoundTrip(req)
	}
	if err := func() error {
		rt.mu.Lock()
		defer rt.mu.Unlock()
		if rt.accessToken == "" || rt.fetchedAt.Add(TokenLifetime).Before(time.Now()) {
			now := time.Now()
			token, err := GetAccessToken(req.Context(), rt.refreshToken)
			if err != nil {
				return err
			}
			rt.accessToken = token.AccessToken
			rt.fetchedAt = now
		}

		if rt.accessToken != "" {
			req.Header.Add("Authorization", "Bearer "+rt.accessToken)
		}
		return nil
	}(); err != nil {
		return nil, err
	}
	return rt.next.RoundTrip(req)
}

func NewRPCClient(endpoint string, refreshToken string) *rpc.Client {
	httpClient := http.Client{
		Timeout: defaultTimeout,
		Transport: newRoundTripper(refreshToken, gzhttp.Transport(&http.Transport{
			MaxIdleConnsPerHost: defaultMaxIdleConnsPerHost,
			DialContext: (&net.Dialer{
				Timeout:   defaultTimeout,
				KeepAlive: defaultKeepAlive,
			}).DialContext,
			ForceAttemptHTTP2:   true,
			TLSHandshakeTimeout: 10 * time.Second,
		})),
	}
	rpcClient := jsonrpc.NewClientWithOpts(endpoint, &jsonrpc.RPCClientOpts{
		HTTPClient: &httpClient,
	})
	return rpc.NewWithCustomRPCClient(rpcClient)
}
