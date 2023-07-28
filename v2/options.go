package v2

import (
	"github.com/fairyhunter13/reflecthelper/v5"
	"github.com/flip-id/valuefirst/manager"
	"github.com/flip-id/valuefirst/storage"
	"github.com/gojek/heimdall/v7"
	"github.com/gojek/heimdall/v7/hystrix"
	"net/http"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the default base URL of Wappin API.
	DefaultBaseURL = "https://api.chat.wappin.app"
	// DefaultTimeout is the default timeout of Wappin API.
	DefaultTimeout = 30 * time.Second
)

// Option is option for initializing Wappin V2 client.
type Option struct {
	BaseURL        string
	LoginURL       string
	MessagesURL    string
	Username       string
	Password       string
	TokenCacheKey  string
	Client         heimdall.Doer
	Timeout        time.Duration
	HystrixOptions []hystrix.Option
	Storage        storage.Hub
	ManagerOptions []manager.FnOption
	client         *hystrix.Client
	wappinClient   *client
}

// Assign assigns the option to the client.
func (o *Option) Assign(opts ...FnOption) *Option {
	for _, opt := range opts {
		opt(o)
	}

	return o
}

// Clone returns a clone of the option.
// Clone only returns a shallow copy of the option.
func (o *Option) Clone() *Option {
	newOpt := *o
	return &newOpt
}

func (o *Option) setWappinClient(c *client) *Option {
	o.wappinClient = c
	return o
}

// Default returns the default option.
func (o *Option) Default() *Option {
	if o.BaseURL == "" {
		o.BaseURL = DefaultBaseURL
	}

	o.BaseURL = strings.TrimRight(o.BaseURL, "/")
	if reflecthelper.IsNil(o.Client) {
		o.Client = http.DefaultClient
	}

	if o.Timeout < DefaultTimeout {
		o.Timeout = DefaultTimeout
	}

	o.client = hystrix.NewClient(
		append(o.HystrixOptions,
			hystrix.WithHTTPTimeout(o.Timeout),
			hystrix.WithHystrixTimeout(o.Timeout),
			hystrix.WithHTTPClient(o.Client),
		)...,
	)

	if o.Storage == nil {
		o.Storage = storage.NewLocalStorage()
	}

	if o.wappinClient == nil {
		o.wappinClient = (new(client)).Assign(o)
	}

	return o
}

// FnOption is a function that modifies an Option
type FnOption func(o *Option)

// WithBaseURL sets the base URL of Wappin API.
func WithBaseURL(baseURL string) FnOption {
	return func(o *Option) {
		o.BaseURL = baseURL
	}
}

// WithLoginURL sets the Login URL of Wappin API.
func WithLoginURL(loginURL string) FnOption {
	return func(o *Option) {
		o.LoginURL = loginURL
	}
}

// WithMessagesURL sets the Messages URL of Wappin API.
func WithMessagesURL(messagesURL string) FnOption {
	return func(o *Option) {
		o.MessagesURL = messagesURL
	}
}

// WithClient sets the client of Wappin API.
func WithClient(client heimdall.Doer) FnOption {
	return func(o *Option) {
		o.Client = client
	}
}

// WithTimeout sets the timeout of Wappin API.
func WithTimeout(timeout time.Duration) FnOption {
	return func(o *Option) {
		o.Timeout = timeout
	}
}

// WithHystrixOptions sets the hystrix options of Wappin API.
func WithHystrixOptions(options ...hystrix.Option) FnOption {
	return func(o *Option) {
		o.HystrixOptions = options
	}
}

// WithStorage sets the token storage of Wappin API.
func WithStorage(storage storage.Hub) FnOption {
	return func(o *Option) {
		o.Storage = storage
	}
}

// WithManagerOptions sets the manager options of Wappin API.
func WithManagerOptions(options ...manager.FnOption) FnOption {
	return func(o *Option) {
		o.ManagerOptions = options
	}
}

// WithTokenCacheKey set token cache key for manage token
func WithTokenCacheKey(tokenCacheKey string) FnOption {
	return func(o *Option) {
		o.TokenCacheKey = tokenCacheKey
	}
}
