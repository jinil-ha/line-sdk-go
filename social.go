package line

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dgrijalva/jwt-go"
)

// UserInfo is ID Token(JWT)'s payload type
type UserInfo struct {
	jwt.StandardClaims

	AuthTime    int    `json:"auth_time"`
	Nonce       string `json:"nonce"`
	DisplayName string `json:"name"`
	PictureURL  string `json:"picture"`
	EMail       string `json:"email,omitempty"`
}

// GetUserInfo get user info from ID Token
func (tk *Token) GetUserInfo() (*UserInfo, error) {
	claims := new(UserInfo)

	jwt, err := jwt.ParseWithClaims(tk.IDToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(tk.channel.channelSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parsing error: %s", err)
	}
	if !jwt.Valid {
		return nil, fmt.Errorf("id token invalid: %v", jwt)
	}

	return claims, nil
}

// Profile is struct of user profile info
type Profile struct {
	DisplayName   string `json:"displayName"`
	UserID        string `json:"userId"`
	PictureURL    string `json:"pictureUrl"`
	StatusMessage string `json:"statusMessage"`
}

// GetProfile get user's profile info from server
func (tk *Token) GetProfile() (*Profile, error) {
	req, err := http.NewRequest("GET", URLProfile, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tk.AccessToken)

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
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%d %s", resp.StatusCode, body)
	}
	//log.Printf("%s", body)

	pf := new(Profile)
	err = json.Unmarshal(body, &pf)
	if err != nil {
		return nil, err
	}
	return pf, nil
}
