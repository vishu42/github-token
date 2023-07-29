package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/lunny/log"
)

type GitHubAppAuth struct {
	AppID             int64
	AppInstallationID int64
	AppPrivateKey     string
}

type AccessToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func createJWTForGitHubApp(appAuth *GitHubAppAuth) (string, error) {
	// Encode as JWT
	// See https://docs.github.com/en/developers/apps/building-github-apps/authenticating-with-github-apps#authenticating-as-a-github-app

	// Going back in time a bit helps with clock skew.
	issuedAt := time.Now().Add(-60 * time.Second)
	// Max expiration date is 10 minutes.
	expiresAt := issuedAt.Add(9 * time.Minute)
	claims := &jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(issuedAt),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		Issuer:    strconv.FormatInt(appAuth.AppID, 10),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(appAuth.AppPrivateKey))
	if err != nil {
		return "", err
	}

	return token.SignedString(privateKey)
}

func FetchAccessToken(ctx context.Context, gitHubConfigURL string, creds *GitHubAppAuth) (*AccessToken, error) {
	accessTokenJWT, err := createJWTForGitHubApp(creds)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/app/installations/%v/access_tokens", creds.AppInstallationID)
	u, err := url.JoinPath(gitHubConfigURL, path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, u, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/vnd.github+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessTokenJWT))

	log.Info("getting access token for GitHub App auth", "accessTokenURL", req.URL.String())

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// if status code is non-200, return error
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("got non-200 status code: %v", resp.StatusCode)
	}

	// Format: https://docs.github.com/en/rest/apps/apps#create-an-installation-access-token-for-an-app
	var accessToken *AccessToken
	err = json.NewDecoder(resp.Body).Decode(&accessToken)

	return accessToken, err
}
