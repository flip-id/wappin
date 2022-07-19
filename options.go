package wappin

import (
	"github.com/fairyhunter13/reflecthelper/v5"
	"github.com/gojek/heimdall/v7"
	"github.com/gojek/heimdall/v7/hystrix"
	"net/http"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the default base URL of Wappin API.
	DefaultBaseURL = "https://api.wappin.id"
	// DefaultTimeout is the default timeout of Wappin API.
	DefaultTimeout = 30 * time.Second
)

// List of all endpoints used in this package.
const (
	EndpointSendHSM = "/v1/message/do-send-hsm"
	EndpointToken   = "/v1/token/get"
)

// Option is option for initializing Wappin client.
type Option struct {
	BaseURL        string
	ClientID       string
	ProjectID      string
	SecretKey      string
	ClientKey      string
	Client         heimdall.Doer
	Timeout        time.Duration
	HystrixOptions []hystrix.Option
	// TODO: Add valuefirst storage hub here.
	// TODO: Add valuefirst token manager options here.
	client *hystrix.Client
}

// Assign assigns the option to the client.
func (o *Option) Assign(opts ...FnOption) *Option {
	for _, opt := range opts {
		opt(o)
	}

	return o
}

// Clone returns a clone of the option.
func (o *Option) Clone() *Option {
	newOpt := *o
	return &newOpt
}

// Default returns the default option.
func (o *Option) Default() *Option {
	if o.BaseURL == "" {
		o.BaseURL = DefaultBaseURL
	}

	o.BaseURL = strings.TrimSuffix(o.BaseURL, "/")
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
