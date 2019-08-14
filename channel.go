package line

import "fmt"

// API URL constants
const (
	URLAuthorize = "https://access.line.me/oauth2/v2.1/authorize"
	URLToken     = "https://api.line.me/oauth2/v2.1/token"
	URLVerify    = "https://api.line.me/oauth2/v2.1/verify"
	URLRevoke    = "https://api.line.me/oauth2/v2.1/revoke"
	URLProfile   = "https://api.line.me/v2/profile"
)

// option strings
const (
	ScopeProfile = "profile"
	ScopeOpenID  = "openid"
	ScopeEmail   = "email"

	PromptConsent = "consent"

	BotPromptNormal     = "normal"
	BotPromptAggressive = "aggressive"
)

// ScopeOpt is type of channel scope option.
// It can be ScopeProfile, ScopeOpenID or ScopeEmail
type ScopeOpt string

// PromptOpt is type of channel prompt option.
// It can be PromptConsent
type PromptOpt string

// MaxAgeOpt is type of channel max age option.
type MaxAgeOpt int

// UILocalesOpt is type of channel UI locale option.
// It currently refers to rfc5646(https://tools.ietf.org/html/rfc5646)
type UILocalesOpt string

// BotPromptOpt is type of channel bot prompt option.
// It can be BotPromptNormal or BotPromptAggressive
type BotPromptOpt string

// Channel type
type Channel struct {
	channelID     string
	channelSecret string
	redirectURI   string

	scope     []string
	prompt    string
	maxAge    int
	uiLocales []string
	botPrompt string
}

// NewChannel return new channel object
func NewChannel(channelID, channelSecret, redirectURI string, opts ...interface{}) (*Channel, error) {
	if channelID == "" {
		return nil, fmt.Errorf("invalid channelID")
	}
	if channelSecret == "" {
		return nil, fmt.Errorf("invalid channelSecret")
	}

	c := &Channel{
		channelID:     channelID,
		channelSecret: channelSecret,
		redirectURI:   redirectURI,

		scope:     []string{},
		prompt:    "",
		maxAge:    0,
		uiLocales: []string{},
		botPrompt: "",
	}

	for _, o := range opts {
		switch v := o.(type) {
		case ScopeOpt:
			c.scope = append(c.scope, string(v))
		case PromptOpt:
			c.prompt = string(v)
		case MaxAgeOpt:
			c.maxAge = int(v)
		case UILocalesOpt:
			c.uiLocales = append(c.uiLocales, string(v))
		case BotPromptOpt:
			c.botPrompt = string(v)
		default:
			return nil, fmt.Errorf("invalid option type")
		}
	}

	// default options
	if len(c.scope) == 0 {
		c.scope = []string{"profile"}
	}

	return c, nil
}
