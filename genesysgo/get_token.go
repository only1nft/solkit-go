package genesysgo

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

type AuthResponse struct {
	AccessToken string `json:"access_token"`
}

func GetAccessToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	if len(refreshToken) == 0 {
		return nil, nil
	}
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"https://auth.genesysgo.net/auth/realms/RPCs/protocol/openid-connect/token",
		strings.NewReader("grant_type=client_credentials"),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic "+refreshToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}
	var data AuthResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	return &data, nil
}
