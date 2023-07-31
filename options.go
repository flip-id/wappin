package wappin

import (
	"net/http"
	"strings"
	"time"

	"github.com/fairyhunter13/reflecthelper/v5"
	"github.com/flip-id/valuefirst/manager"
	"github.com/flip-id/valuefirst/storage"
	"github.com/gojek/heimdall/v7"
	"github.com/gojek/heimdall/v7/hystrix"
)

const (
	// DefaultBaseURL is the default base URL of Wappin API.
	DefaultBaseURL = "https://api.wappin.id"
	// DefaultTimeout is the default timeout of Wappin API.
	DefaultTimeout = 30 * time.Second
	// 	DefaultTokenKey is the default key for the token storage.
	DefaultTokenKey = "manager:token:wappin"
	// DefaultMarketingTokenKey is the default key for the token storage especially for marketing account.
	DefaultMarketingTokenKey = "manager:marketing-token:wappin"
)

// List of all endpoints used in this package.
const (
	EndpointSendHSM = "/v1/message/do-send-hsm"
	EndpointToken   = "/v1/token/get"
)

// Option is option for initializing Wappin client.
type Option struct {
	BaseURL            string
	ClientID           string
	ProjectID          string
	SecretKey          string
	ClientKey          string
	TokenCacheKey      string
	Client             heimdall.Doer
	Timeout            time.Duration
	HystrixOptions     []hystrix.Option
	Storage            storage.Hub
	ManagerOptions     []manager.FnOption
	client             *hystrix.Client
	wappinClient       *client
	manager            manager.TokenManager
	IsMarketingAccount bool
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

	// The default key for SendOTP account
	key := DefaultTokenKey
	// This key for marketing account
	if o.IsMarketingAccount {
		key = DefaultMarketingTokenKey
	}

	// The new config for any account you want to use, so this config will be used in the future
	if o.TokenCacheKey != "" {
		key = o.TokenCacheKey
	} else {
		// just to help create a new redis cache key for generate new token
		o.TokenCacheKey = key
	}

	o.manager = manager.New(
		append([]manager.FnOption{
			manager.WithStorage(o.Storage),
			manager.WithClient(o.wappinClient),
			manager.WithKey(key),
		}, o.ManagerOptions...)...,
	)

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

// WithClientID sets the client ID of Wappin API.
func WithClientID(clientID string) FnOption {
	return func(o *Option) {
		o.ClientID = clientID
	}
}

// WithProjectID sets the project ID of Wappin API.
func WithProjectID(projectID string) FnOption {
	return func(o *Option) {
		o.ProjectID = projectID
	}
}

// WithSecretKey sets the secret key of Wappin API.
func WithSecretKey(secretKey string) FnOption {
	return func(o *Option) {
		o.SecretKey = secretKey
	}
}

// WithClientKey sets the client key of Wappin API.
func WithClientKey(clientKey string) FnOption {
	return func(o *Option) {
		o.ClientKey = clientKey
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

// WithIsMarketingAccount sets Is Marketing account or not.
func WithIsMarketingAccount(isMarketingAccount bool) FnOption {
	return func(o *Option) {
		o.IsMarketingAccount = isMarketingAccount
	}
}

// WithTokenCacheKey set token cache key for manage token
func WithTokenCacheKey(tokenCacheKey string) FnOption {
	return func(o *Option) {
		o.TokenCacheKey = tokenCacheKey
	}
}
