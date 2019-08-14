package line

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const randomLetters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

const (
	stateLength = 32
	nonceLength = 32
)

// Login type
type Login struct {
	Channel *Channel

	state string
	nonce string

	FriendshipStatusChanged bool
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randomString(length int) string {
	b := make([]byte, length)
	l := len(randomLetters)

	for i := range b {
		b[i] = randomLetters[rand.Intn(l)]
	}
	return string(b)
}

// NewLogin return url for LINE login
func (c *Channel) NewLogin() *Login {
	l := new(Login)
	l.Channel = c
	l.state = randomString(stateLength)
	l.nonce = randomString(nonceLength)

	return l
}

// AuthorizeURL return authorize url of a session for LINE Login
func (l *Login) AuthorizeURL() string {
	authURL, _ := url.Parse(URLAuthorize)

	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", l.Channel.channelID)
	params.Add("redirect_uri", l.Channel.redirectURI)
	params.Add("state", l.state)
	params.Add("scope", strings.Join(l.Channel.scope, " "))
	params.Add("nonce", l.nonce)
	if l.Channel.prompt != "" {
		params.Add("prompt", l.Channel.prompt)
	}
	if l.Channel.maxAge > 0 {
		params.Add("max_age", string(l.Channel.maxAge))
	}
	if len(l.Channel.uiLocales) > 0 {
		params.Add("ui_locales", strings.Join(l.Channel.uiLocales, " "))
	}
	if l.Channel.botPrompt != "" {
		params.Add("bot_prompt", l.Channel.botPrompt)
	}

	authURL.RawQuery = params.Encode()
	return authURL.String()
}

// Auth implements LINE auth.
// param is param info redirected from LINE.
// and use ID Token as token
func (l *Login) Auth(r *http.Request) (*Token, error) {
	q := r.URL.Query()
	code := q.Get("code")
	state := q.Get("state")
	changed := q.Get("friendship_status_changed")
	error := q.Get("error")
	desc := q.Get("error_description")

	if error != "" {
		return nil, fmt.Errorf("%s %s", error, desc)
	}

	log.Printf("login: %v", l)
	if l.state != state {
		return nil, fmt.Errorf("Invalid state: sent(%s) received(%s)", l.state, state)
	}
	l.FriendshipStatusChanged = changed == "true"

	// get line token
	return l.Channel.GetToken(code)
}
