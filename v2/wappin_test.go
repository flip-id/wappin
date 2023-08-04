package v2_test

import (
	"context"
	"fmt"
	v2 "github.com/flip-id/wappin/v2"
	"github.com/go-redis/redis/v8"
	"github.com/gojek/heimdall/v7/hystrix"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gojek/valkyrie"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	wappinToken            = "qwertyuiopsasdfghjklzxcvbnm"
	tokenCacheKeyMarketing = "manager:marketing:token_wappin_v2"

	successLoginResponseJson       = `{"meta":{"version":"1.0.4"},"users":[{"token":"eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6IjY0N2NiOGFhZGIzN2FhNzg4NmU1OWYxMSIsImlhdCI6IjE2OTEwMzQzMzYiLCJlYXQiOiIxNjkxMDM0MzM2IiwidWlkIjoiYTA1OTE0ZDQtZWIyOC00MWY4LTkzZDctODBjNmE0MWUwOTlhIn0.yWecKEL-4lEySEJ-5w84N3m7IIK5OJdIfQs8n659lzY","expired_after":"2023-08-03T10:45:36+07:00"}]}`
	successSendMessageResponseJson = `{"meta":{"version":"1.0.4"},"messages":[{"id":"FkV4YbHb_X8sfVYcPQ7_QEpK_Rvc"}]}`
	errorLoginResponseJson         = `{"meta":{"version":"1.0.4"},"errors":[{"code":1005,"title":"Access Denied","details":"Missing or invalid authentication credentials."}]}`
	errorGeneralResponseJson       = `{"meta":{"version":"1.0.4"},"errors":[{"code":500,"title":"General Error","details":"General Error"}]}`

	invalidCredentialErr = v2.CastError(1005, "Access Denied", "Missing or invalid authentication credentials.")
	generalErr           = v2.CastError(500, "General Error", "General Error")

	requestSendMessage = v2.RequestMessage{
		To:   "6288889999",
		Type: v2.MessageTypeTemplate,
		Template: v2.TemplateRequest{
			Name: "testing_webhook_marketing",
			Language: v2.LanguageRequest{
				Policy: "deterministic",
				Code:   "id",
			},
			Namespace: "9898912-121212",
			Components: []v2.ComponentRequest{
				{
					Type: v2.ComponentTypeBody,
					Parameters: []v2.ComponentParameterRequest{
						{
							Type: v2.MessageTypeText,
							Text: "hari ini",
						},
						{
							Type: v2.MessageTypeText,
							Text: "Rp. 999999999",
						},
					},
				},
			},
		},
	}

	responseSuccessSendMessage = v2.ResponseMessage{
		BaseResponse: v2.BaseResponse{
			Meta: v2.MetaResponse{
				Version: "1.0.4",
			},
			Errors: nil,
		},
		Messages: []v2.MessageResponse{
			{Id: "FkV4YbHb_X8sfVYcPQ7_QEpK_Rvc"},
		},
	}
)

type (
	response struct {
		status       int
		jsonResponse string
	}

	doerMock struct {
		DoLoginFunc    func(*http.Request) (*response, error)
		DoMessagesFunc func(*http.Request) (*response, error)
	}

	storageMock struct {
		GetFunc  func(ctx context.Context, key string) (i interface{}, err error)
		SaveFunc func(ctx context.Context, key string, i interface{}, ttl time.Duration) (err error)
	}
)

func (s storageMock) Get(ctx context.Context, key string) (i interface{}, err error) {
	getFunc, err := s.GetFunc(ctx, key)
	if err != nil {
		return nil, err
	}

	return getFunc, nil
}

func (s storageMock) Save(ctx context.Context, key string, i interface{}, ttl time.Duration) (err error) {
	err = s.SaveFunc(ctx, key, i, ttl)
	if err != nil {
		return
	}

	return nil
}

func (d *doerMock) Do(req *http.Request) (*http.Response, error) {
	url := req.URL.String()
	status := http.StatusOK
	var jsonResponse string

	if url == "https://base_url/v1/users/login" {
		// jsonResponse = defaultJsonResponse
		if d.DoLoginFunc != nil {
			r, err := d.DoLoginFunc(req)
			if err != nil {
				return nil, err
			}

			status = r.status
			jsonResponse = r.jsonResponse
		}
	}

	if url == "https://base_url/v1/messages" {
		if d.DoMessagesFunc != nil {
			r, err := d.DoMessagesFunc(req)
			if err != nil {
				return nil, err
			}

			status = r.status
			jsonResponse = r.jsonResponse
		}
	}

	return &http.Response{
		Status:     fmt.Sprintf("%d", status),
		StatusCode: status,
		Header: map[string][]string{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(jsonResponse)),
	}, nil
}

type wappinTestSuite struct {
	suite.Suite

	wp            v2.Client
	doer          *doerMock
	storageMock   storageMock
	tokenCacheKey string
}

func TestWappinTestSuite(t *testing.T) {
	ts := wappinTestSuite{}

	ts.tokenCacheKey = tokenCacheKeyMarketing

	suite.Run(t, &ts)
}

func (ts *wappinTestSuite) TestSendMessage() {
	tt := []struct {
		name      string
		args      func() *v2.RequestMessage
		mock      func(r *v2.RequestMessage)
		expect    func() (*v2.ResponseMessage, error)
		expectErr bool
	}{
		{
			name: "Success send message",
			args: func() *v2.RequestMessage {
				return &requestSendMessage
			},
			mock: func(r *v2.RequestMessage) {
				ts.doer = &doerMock{
					DoLoginFunc: func(r *http.Request) (*response, error) {
						return &response{
							status:       401,
							jsonResponse: successLoginResponseJson,
						}, nil
					},
					DoMessagesFunc: func(request *http.Request) (*response, error) {
						return &response{
							status:       200,
							jsonResponse: successSendMessageResponseJson,
						}, nil
					},
				}

				ts.storageMock = storageMock{
					GetFunc: func(ctx context.Context, key string) (i interface{}, err error) {
						return nil, redis.Nil
					},
					SaveFunc: func(ctx context.Context, key string, i interface{}, ttl time.Duration) (err error) {
						return nil
					},
				}

				ts.wp = v2.New(
					v2.WithHystrixOptions(hystrix.WithErrorPercentThreshold(100)),
					v2.WithClient(ts.doer),
					v2.WithStorage(ts.storageMock),
					v2.WithBaseURL("https://base_url"),
					v2.WithLoginURL("/v1/users/login"),
					v2.WithMessagesURL("/v1/messages"),
				)
			},
			expect: func() (*v2.ResponseMessage, error) {
				return &responseSuccessSendMessage, nil
			},
			expectErr: false,
		},
		{
			name: "Success send message with get token from cache",
			args: func() *v2.RequestMessage {
				return &requestSendMessage
			},
			mock: func(r *v2.RequestMessage) {
				ts.doer = &doerMock{
					DoMessagesFunc: func(request *http.Request) (*response, error) {
						return &response{
							status:       200,
							jsonResponse: successSendMessageResponseJson,
						}, nil
					},
				}

				ts.storageMock = storageMock{
					GetFunc: func(ctx context.Context, key string) (i interface{}, err error) {
						return wappinToken, nil
					},
					SaveFunc: func(ctx context.Context, key string, i interface{}, ttl time.Duration) (err error) {
						return nil
					},
				}

				ts.wp = v2.New(
					v2.WithHystrixOptions(hystrix.WithErrorPercentThreshold(100)),
					v2.WithClient(ts.doer),
					v2.WithStorage(ts.storageMock),
					v2.WithBaseURL("https://base_url"),
					v2.WithLoginURL("/v1/users/login"),
					v2.WithMessagesURL("/v1/messages"),
				)
			},
			expect: func() (*v2.ResponseMessage, error) {
				return &responseSuccessSendMessage, nil
			},
			expectErr: false,
		},
		{
			name: "Error login invalid credential from Wappin",
			args: func() *v2.RequestMessage {
				return &requestSendMessage
			},
			mock: func(r *v2.RequestMessage) {
				ts.doer = &doerMock{
					DoLoginFunc: func(r *http.Request) (*response, error) {
						return &response{
							status:       401,
							jsonResponse: errorLoginResponseJson,
						}, nil
					},
				}

				ts.storageMock = storageMock{
					GetFunc: func(ctx context.Context, key string) (i interface{}, err error) {
						return nil, redis.Nil
					},
					SaveFunc: func(ctx context.Context, key string, i interface{}, ttl time.Duration) (err error) {
						return nil
					},
				}

				ts.wp = v2.New(
					v2.WithHystrixOptions(hystrix.WithErrorPercentThreshold(100)),
					v2.WithClient(ts.doer),
					v2.WithStorage(ts.storageMock),
					v2.WithBaseURL("https://base_url"),
					v2.WithLoginURL("/v1/users/login"),
					v2.WithMessagesURL("/v1/messages"),
				)
			},
			expect: func() (*v2.ResponseMessage, error) {
				return nil, invalidCredentialErr
			},
			expectErr: true,
		},
		{
			name: "Error send message general from Wappin",
			args: func() *v2.RequestMessage {
				return &requestSendMessage
			},
			mock: func(r *v2.RequestMessage) {
				ts.doer = &doerMock{
					DoLoginFunc: func(r *http.Request) (*response, error) {
						return &response{
							status:       200,
							jsonResponse: successLoginResponseJson,
						}, nil
					},
					DoMessagesFunc: func(request *http.Request) (*response, error) {
						return &response{
							status:       500,
							jsonResponse: errorGeneralResponseJson,
						}, nil
					},
				}

				ts.storageMock = storageMock{
					GetFunc: func(ctx context.Context, key string) (i interface{}, err error) {
						return nil, redis.Nil
					},
					SaveFunc: func(ctx context.Context, key string, i interface{}, ttl time.Duration) (err error) {
						return nil
					},
				}

				ts.wp = v2.New(
					v2.WithHystrixOptions(hystrix.WithErrorPercentThreshold(100)),
					v2.WithClient(ts.doer),
					v2.WithStorage(ts.storageMock),
					v2.WithBaseURL("https://base_url"),
					v2.WithLoginURL("/v1/users/login"),
					v2.WithMessagesURL("/v1/messages"),
				)
			},
			expect: func() (*v2.ResponseMessage, error) {
				return nil, generalErr
			},
			expectErr: true,
		},
		{
			name: "Error nil arguments",
			args: func() *v2.RequestMessage {
				return nil
			},
			mock: func(r *v2.RequestMessage) {
				ts.doer = &doerMock{}

				ts.storageMock = storageMock{}

				ts.wp = v2.New(
					v2.WithHystrixOptions(hystrix.WithErrorPercentThreshold(100)),
					v2.WithClient(ts.doer),
					v2.WithStorage(ts.storageMock),
					v2.WithBaseURL("https://base_url"),
					v2.WithLoginURL("/v1/users/login"),
					v2.WithMessagesURL("/v1/messages"),
				)
			},
			expect: func() (*v2.ResponseMessage, error) {
				return nil, errors.New("Request nil arguments")
			},
			expectErr: true,
		},
	}

	for _, tc := range tt {
		ts.T().Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			args := tc.args()
			tc.mock(args)

			expectResp, expectErr := tc.expect()

			response, err := ts.wp.SendMessage(ctx, args)
			if tc.expectErr {
				mErr, ok := err.(*valkyrie.MultiError)
				if ok {
					err = mErr
				}

				assert.Equal(t, expectResp, response)
				assert.Equal(t, expectErr.Error(), err.Error())

				return
			}

			if err != nil {
				assert.Equal(t, expectErr.Error(), err.Error())
			}
			assert.Equal(t, expectResp, response)
		})
	}

}
