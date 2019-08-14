package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	line "github.com/jinil-ha/line-sdk-go"
)

const (
	homeHTML = `<a href="/goauth">[LINE Login]</a>`
	topHTML  = `<html><head>
<meta http-equiv="Cache-Control" content="no-cache, no-store, must-revalidate">
<meta http-equiv="Pragma" content="no-cache">
<meta http-equiv="Expires" content="0"></head>
<body>[ID Token(JWT) Info]<br/>
DisplayName : %s<br/>E-Mail : %s<br/>
<img src="%s" width=100 height=100/><br/>
[Profile Info]<br/>
ID : %s<br/>Name : %s<br/>Status : %s<br/>
<img src="%s" width=100 height=100/><br/>
[Token Info]<br/>
Access Token : %s...<br/>Refresh Token : %s...<br/>Id Token : %s...<br/>ExpireTime : %s<br/><br/>
<a href="/verify">[Verify]</a><br/>
<a href="/refresh">[Refresh]</a><br/>
<a href="/revoke">[Revoke]</a><br/>
<a href="/logout">[Logout]</a><br/>
</body></html>`
	tokenHTML = `<body>%s<br/><a href="/">[Home]</a></body>`
)

// This is sample server for SINGLE user
func main() {
	var login *line.Login
	var token *line.Token

	channel, err := line.NewChannel(os.Getenv("CHANNEL_ID"),
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("REDIRECT_URI"),
		line.ScopeOpt(line.ScopeProfile),
		line.ScopeOpt(line.ScopeOpenID),
		line.BotPromptOpt(line.BotPromptAggressive))
	if err != nil {
		log.Printf("Channel create error: %s", err)
		os.Exit(1)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if token == nil {
			fmt.Fprintf(w, homeHTML)
			return
		}

		log.Printf("Enter: token(%v)\n", token)

		user, err := token.GetUserInfo()
		if err != nil {
			log.Printf("Get user info failed: %s", err)
			// http.Redirect(w, r, "/logout", 301)
			// return
		}

		pf, err := token.GetProfile()
		if err != nil {
			log.Printf("Get profile failed: %s", err)
			pf = new(line.Profile)
		}

		fmt.Fprintf(w, topHTML,
			user.DisplayName, user.EMail, user.PictureURL,
			pf.UserID, pf.DisplayName, pf.StatusMessage, pf.PictureURL,
			token.AccessToken[:8], token.RefreshToken[:8], token.IDToken[:8],
			token.ExpireTime.Format("01/02 15:04:05"))
	})

	http.HandleFunc("/goauth", func(w http.ResponseWriter, r *http.Request) {
		login = channel.NewLogin()
		url := login.AuthorizeURL()
		log.Printf("LINE login start: %v\n", login)

		w.Header().Set("Cache-Control", "no-cache")
		http.Redirect(w, r, url, 301)
	})

	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("LINE login end: %v\n", login)
		token, err = login.Auth(r)
		if err != nil {
			log.Printf("LINR Login failed: %s", err)
		} else {
			log.Printf("LINE Login OK: %v", token)
		}
		login = nil
		http.Redirect(w, r, "/", 301)
	})

	http.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		if token == nil {
			log.Printf("Verify failed: not logined")
			http.Redirect(w, r, "/", 301)
			return
		}

		log.Printf("verify token: %v", token)
		if err = token.Verify(); err != nil {
			fmt.Fprintf(w, tokenHTML, "Fail to verify")
			log.Printf("Fail to verify: %s acc(%s)", err, token.AccessToken)
		} else {
			fmt.Fprintf(w, tokenHTML, "Succeeded to verify.")
			log.Printf("Access token verified: (%s)", token.AccessToken)
		}
	})

	http.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		if token == nil {
			log.Printf("Refresh failed: not logined")
			http.Redirect(w, r, "/", 301)
			return
		}

		log.Printf("refresh token: %v", token)
		err = token.Refresh()
		if err != nil {
			fmt.Fprintf(w, tokenHTML, "Fail to refresh")
			log.Printf("Fail to refresh: %s ref(%s)", err, token.RefreshToken)
		} else {
			fmt.Fprintf(w, tokenHTML, "Refresh OK")
			log.Printf("Access token refreshed: (%s)", token.AccessToken)
		}
	})

	http.HandleFunc("/revoke", func(w http.ResponseWriter, r *http.Request) {
		if token == nil {
			log.Printf("Revoke failed: not logined")
			http.Redirect(w, r, "/", 301)
			return
		}

		log.Printf("revoke token: %v", token)
		err = token.Revoke()
		if err != nil {
			fmt.Fprintf(w, tokenHTML, "Fail to revoke")
			log.Printf("Fail to revoke: %s acc(%s)", err, token.AccessToken)
		} else {
			fmt.Fprintf(w, tokenHTML, "Revoke OK")
			log.Printf("Access token revoked: (%s)", token.AccessToken)
		}
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Logout: %v", token)
		login = nil
		token = nil
		w.Header().Set("Cache-Control", "no-cache")
		http.Redirect(w, r, "/", 301)
	})

	http.ListenAndServe(":80", nil)
}
