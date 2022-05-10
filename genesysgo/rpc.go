package genesysgo

import (
	"context"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
	"github.com/klauspost/compress/gzhttp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

var (
	TokenURL                   = "https://auth.genesysgo.net/auth/realms/RPCs/protocol/openid-connect/token"
	defaultMaxIdleConnsPerHost = 20
	defaultTimeout             = 5 * time.Minute
	defaultKeepAlive           = 3 * time.Minute

	MainNetBeta = rpc.Cluster{
		Name: rpc.MainNetBeta.Name,
		RPC:  "https://only1.genesysgo.net",
		WS:   "ws://only1.genesysgo.net",
	}

	DevNet = rpc.Cluster{
		Name: rpc.DevNet.Name,
		RPC:  "https://psytrbhymqlkfrhudd.dev.genesysgo.net:8899",
		WS:   "ws://psytrbhymqlkfrhudd.dev.genesysgo.net:8899",
	}

	ErrInvalidToken = errors.New("invalid token")
)

type roundTripper struct {
	clientcredentials.Config
	next http.RoundTripper
}

func newRoundTripper(cfg clientcredentials.Config, next http.RoundTripper) *roundTripper {
	out := roundTripper{next: next, Config: cfg}
	return &out
}

func (rt *roundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	token, err := rt.Token(req.Context())
	if err != nil {
		return nil, err
	}
	token.SetAuthHeader(req)
	return rt.next.RoundTrip(req)
}

func decodeAuthToken(b64Token string) (*clientcredentials.Config, error) {
	if b64Token == "" {
		return nil, nil
	}
	decoded, err := base64.StdEncoding.DecodeString(b64Token)
	if err != nil {
		return nil, ErrInvalidToken
	}
	s := strings.SplitN(string(decoded), ":", 2)
	if len(s) != 2 {
		return nil, ErrInvalidToken
	}
	return &clientcredentials.Config{
		ClientID:     s[0],
		TokenURL:     TokenURL,
		ClientSecret: s[1],
	}, nil
}

// RFC2617
// authToken - clientID:clientSecret encoded in base64
// Please use the following instruction in order to generate a new token:
// https://genesysgo.medium.com/a-primer-to-genesysgo-network-auth-a3c678a9dc2a
func GetToken(ctx context.Context, authToken string) (*oauth2.Token, error) {
	cfg, err := decodeAuthToken(authToken)
	if err != nil {
		return nil, err
	}
	return cfg.Token(ctx)
}

// RFC2617
// authToken - clientID:clientSecret encoded in base64
// Please use the following instruction in order to generate a new token:
// https://genesysgo.medium.com/a-primer-to-genesysgo-network-auth-a3c678a9dc2a
func NewRPCClient(endpoint string, authToken string) *rpc.Client {
	if authToken == "" {
		return rpc.New(endpoint)
	}
	cfg, err := decodeAuthToken(authToken)
	if err != nil {
		panic(err)
	}
	httpClient := http.Client{
		Timeout: defaultTimeout,
		Transport: newRoundTripper(*cfg, gzhttp.Transport(&http.Transport{
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
