package line

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// request parameter string of "grant_type"
const (
	GrantTypeGet     = "authorization_code"
	GrantTypeRefresh = "refresh_token"
)

// Token is struct of tokens from LINE Login
type Token struct {
	AccessToken  string
	IDToken      string
	RefreshToken string
	ExpireTime   time.Time

	channel *Channel
}

// tokenResp is struct of response from LINE auth server
type tokenResp struct {
	AccessToken  string `json:"access_token,omitempty"`
	ExpiresIn    int64  `json:"expires_in,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type,omitempty"`
	ClientID     string `json:"client_id,omitempty"`
}

// NewToken return new token object with access token and id token.
// idToken is not necessary.
// if idToken is not passed(""), it will be failed to get User Info with GetUserInfo().
func (c *Channel) NewToken(accessToken string, idToken string) (*Token, error) {
	token := new(Token)
	token.AccessToken = accessToken
	token.IDToken = idToken

	return token, nil
}

// GetToken get access token from LINE oauth
func (c *Channel) GetToken(code string) (*Token, error) {
	v := url.Values{}
	v.Add("grant_type", GrantTypeGet)
	v.Add("code", code)
	v.Add("redirect_uri", c.redirectURI)
	v.Add("client_id", c.channelID)
	v.Add("client_secret", c.channelSecret)
	req, err := http.NewRequest("POST", URLToken, strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tr tokenResp
	err = json.Unmarshal(body, &tr)
	if err != nil {
		return nil, err
	}

	token := new(Token)
	token.AccessToken = tr.AccessToken
	token.IDToken = tr.IDToken
	token.RefreshToken = tr.RefreshToken
	token.ExpireTime = time.Now().Add(time.Duration(tr.ExpiresIn * 1000000000))
	token.channel = c
	return token, nil
}

// Verify request to verify access token
func (tk *Token) Verify() error {
	req, err := http.NewRequest("GET", URLVerify, nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("access_token", tk.AccessToken)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("%d %s", resp.StatusCode, body)
	}

	var tr tokenResp
	err = json.Unmarshal(body, &tr)
	if err != nil {
		return err
	}
	if tr.ClientID != tk.channel.channelID {
		return fmt.Errorf("Invalid channel ID")
	}
	tk.ExpireTime = time.Now().Add(time.Duration(tr.ExpiresIn * 1000000000))
	return nil
}

// Refresh request to refresh using refresh token
func (tk *Token) Refresh() error {
	v := url.Values{}
	v.Add("grant_type", GrantTypeRefresh)
	v.Add("refresh_token", tk.RefreshToken)
	v.Add("client_id", tk.channel.channelID)
	v.Add("client_secret", tk.channel.channelSecret)
	req, err := http.NewRequest("POST", URLToken, strings.NewReader(v.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var tr tokenResp
	err = json.Unmarshal(body, &tr)
	if err != nil {
		return err
	}

	tk.AccessToken = tr.AccessToken
	//tk.RefreshToken = tr.RefreshToken
	tk.ExpireTime = time.Now().Add(time.Duration(tr.ExpiresIn * 1000000000))
	return nil
}

// Revoke request to revoke access token
func (tk *Token) Revoke() error {
	v := url.Values{}
	v.Add("access_token", tk.AccessToken)
	v.Add("client_id", tk.channel.channelID)
	v.Add("client_secret", tk.channel.channelSecret)
	req, err := http.NewRequest("POST", URLRevoke, strings.NewReader(v.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("server returns %d", resp.StatusCode)
	}
	return nil
}
