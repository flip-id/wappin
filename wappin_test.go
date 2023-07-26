package wappin_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/flip-id/wappin"
	"github.com/gojek/heimdall/v7/hystrix"
	"github.com/gojek/valkyrie"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	defaultJsonResponse = `{"status":"200","message":"","data":{"access_token":"access-token","expired_datetime":"2077-12-31 00:00:00","token_type":""}}`
)

type (
	response struct {
		status       int
		jsonResponse string
	}

	doerMock struct {
		SendMessageCalledTimes int
		DoGenerateTokenFunc    func(*http.Request) (*response, error)
		DoSendMessageFunc      func(*http.Request) (*response, error)
	}
)

func (d *doerMock) Do(req *http.Request) (*http.Response, error) {
	url := req.URL.String()
	status := http.StatusOK
	var jsonResponse string

	if url == "https://api.wappin.id/v1/token/get" {
		// jsonResponse = defaultJsonResponse
		if d.DoGenerateTokenFunc != nil {
			r, err := d.DoGenerateTokenFunc(req)
			if err != nil {
				return nil, err
			}

			status = r.status
			jsonResponse = r.jsonResponse
		}
	}

	if url == "https://api.wappin.id/v1/message/do-send-hsm" {
		d.SendMessageCalledTimes += 1
		// jsonResponse = defaultJsonResponse
		if d.DoSendMessageFunc != nil {
			r, err := d.DoSendMessageFunc(req)
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

	wp   wappin.Client
	doer *doerMock
}

func TestWappinTestSuite(t *testing.T) {
	suite.Run(t, &wappinTestSuite{})
}

func (ts *wappinTestSuite) TestSendMessage() {
	tt := []struct {
		name                   string
		args                   func() *wappin.RequestWhatsappMessage
		mock                   func(r *wappin.RequestWhatsappMessage)
		expect                 func() (*wappin.ResponseMessage, error)
		expectErr              bool
		sendMessageCalledTimes int
	}{
		{
			name: "error token manager get - wappin error",
			args: func() *wappin.RequestWhatsappMessage {
				return &wappin.RequestWhatsappMessage{
					ClientID:        "client-id",
					ProjectID:       "project-id",
					Type:            "type",
					RecipientNumber: "081213141516",
					Token:           "token",
				}
			},
			mock: func(r *wappin.RequestWhatsappMessage) {
				ts.doer = &doerMock{
					DoGenerateTokenFunc: func(r *http.Request) (*response, error) {
						return nil, &wappin.Error{
							Status: "400",
						}
					},
				}
				ts.wp = wappin.New(
					wappin.WithHystrixOptions(hystrix.WithErrorPercentThreshold(100)),
					wappin.WithClient(ts.doer),
				)
			},
			expect: func() (*wappin.ResponseMessage, error) {
				return nil, &wappin.Error{
					Status: "400",
				}
			},
			expectErr:              true,
			sendMessageCalledTimes: 0,
		},
		{
			name: "error token manager get - wappin error 2",
			args: func() *wappin.RequestWhatsappMessage {
				return &wappin.RequestWhatsappMessage{
					ClientID:        "client-id",
					ProjectID:       "project-id",
					Type:            "type",
					RecipientNumber: "081213141516",
					Token:           "token",
				}
			},
			mock: func(r *wappin.RequestWhatsappMessage) {
				ts.doer = &doerMock{
					DoGenerateTokenFunc: func(r *http.Request) (*response, error) {
						return &response{
							status:       http.StatusBadRequest,
							jsonResponse: `{"status":"400","message":"error generate token bad request"}`,
						}, nil
					},
				}
				ts.wp = wappin.New(
					wappin.WithHystrixOptions(hystrix.WithErrorPercentThreshold(100)),
					wappin.WithClient(ts.doer),
				)
			},
			expect: func() (*wappin.ResponseMessage, error) {
				return nil, &wappin.Error{
					Status:  "400",
					Message: "error generate token bad request",
				}
			},
			expectErr:              true,
			sendMessageCalledTimes: 0,
		},
		{
			name: "error token manager get - non wappin error",
			args: func() *wappin.RequestWhatsappMessage {
				return &wappin.RequestWhatsappMessage{
					ClientID:        "client-id",
					ProjectID:       "project-id",
					Type:            "type",
					RecipientNumber: "081213141516",
					Token:           "token",
				}
			},
			mock: func(r *wappin.RequestWhatsappMessage) {
				ts.doer = &doerMock{
					DoGenerateTokenFunc: func(r *http.Request) (*response, error) {
						return nil, errors.New("non wappin error")
					},
				}
				ts.wp = wappin.New(
					wappin.WithHystrixOptions(hystrix.WithErrorPercentThreshold(100)),
					wappin.WithClient(ts.doer),
				)
			},
			expect: func() (*wappin.ResponseMessage, error) {
				return nil, errors.New("non wappin error")
			},
			expectErr:              true,
			sendMessageCalledTimes: 0,
		},
		{
			name: "error do - wappin error",
			args: func() *wappin.RequestWhatsappMessage {
				return &wappin.RequestWhatsappMessage{
					ClientID:        "client-id",
					ProjectID:       "project-id",
					Type:            "type",
					RecipientNumber: "081213141516",
					Token:           "token",
				}
			},
			mock: func(r *wappin.RequestWhatsappMessage) {
				ts.doer = &doerMock{
					DoGenerateTokenFunc: func(r *http.Request) (*response, error) {
						return &response{status: http.StatusOK, jsonResponse: defaultJsonResponse}, nil
					},
					DoSendMessageFunc: func(r *http.Request) (*response, error) {
						return &response{
							status:       http.StatusBadRequest,
							jsonResponse: `{"status":"400","message":"error bad request"}`,
						}, nil
					},
				}
				ts.wp = wappin.New(
					wappin.WithHystrixOptions(hystrix.WithErrorPercentThreshold(100)),
					wappin.WithClient(ts.doer),
				)
			},
			expect: func() (*wappin.ResponseMessage, error) {
				return &wappin.ResponseMessage{
						Status:         "400",
						Message:        "error bad request",
						HttpStatusCode: http.StatusBadRequest,
						RawData:        `{"status":"400","message":"error bad request"}`,
					}, &wappin.Error{
						Status:  "400",
						Message: "error bad request",
					}
			},
			expectErr:              true,
			sendMessageCalledTimes: 1,
		},
		{
			name: "error do - non wappin error",
			args: func() *wappin.RequestWhatsappMessage {
				return &wappin.RequestWhatsappMessage{
					ClientID:        "client-id",
					ProjectID:       "project-id",
					Type:            "type",
					RecipientNumber: "081213141516",
					Token:           "token",
				}
			},
			mock: func(r *wappin.RequestWhatsappMessage) {
				ts.doer = &doerMock{
					DoGenerateTokenFunc: func(r *http.Request) (*response, error) {
						return &response{status: http.StatusOK, jsonResponse: defaultJsonResponse}, nil
					},
					DoSendMessageFunc: func(r *http.Request) (*response, error) {
						return nil, errors.New("error do non wappin")
					},
				}
				ts.wp = wappin.New(
					wappin.WithHystrixOptions(hystrix.WithErrorPercentThreshold(100)),
					wappin.WithClient(ts.doer),
				)
			},
			expect: func() (*wappin.ResponseMessage, error) {
				return nil, errors.New("error do non wappin")
			},
			expectErr:              true,
			sendMessageCalledTimes: 1,
		},
		{
			name: "success do",
			args: func() *wappin.RequestWhatsappMessage {
				return &wappin.RequestWhatsappMessage{
					ClientID:        "client-id",
					ProjectID:       "project-id",
					Type:            "type",
					RecipientNumber: "081213141516",
					Token:           "token",
				}
			},
			mock: func(r *wappin.RequestWhatsappMessage) {
				ts.doer = &doerMock{
					DoGenerateTokenFunc: func(r *http.Request) (*response, error) {
						return &response{status: http.StatusOK, jsonResponse: defaultJsonResponse}, nil
					},
					DoSendMessageFunc: func(r *http.Request) (*response, error) {
						return &response{
							status:       http.StatusOK,
							jsonResponse: `{"status":"200","message":"send message success"}`,
						}, nil
					},
				}
				ts.wp = wappin.New(
					wappin.WithHystrixOptions(hystrix.WithErrorPercentThreshold(100)),
					wappin.WithClient(ts.doer),
				)
			},
			expect: func() (*wappin.ResponseMessage, error) {
				return &wappin.ResponseMessage{
					Status:         "200",
					Message:        "send message success",
					HttpStatusCode: http.StatusOK,
					RawData:        `{"status":"200","message":"send message success"}`,
				}, nil
			},
			expectErr:              false,
			sendMessageCalledTimes: 1,
		},
		{
			name: "success - wappin error 401",
			args: func() *wappin.RequestWhatsappMessage {
				return &wappin.RequestWhatsappMessage{
					ClientID:        "client-id",
					ProjectID:       "project-id",
					Type:            "type",
					RecipientNumber: "081213141516",
					Token:           "token",
				}
			},
			mock: func(r *wappin.RequestWhatsappMessage) {
				ts.doer = &doerMock{
					DoGenerateTokenFunc: func(r *http.Request) (*response, error) {
						return &response{status: http.StatusOK, jsonResponse: defaultJsonResponse}, nil
					},
					DoSendMessageFunc: func(r *http.Request) (*response, error) {
						return &response{
							status:       http.StatusUnauthorized,
							jsonResponse: `{"status":"401","message":"error invalid token"}`,
						}, nil
					},
				}
				ts.wp = wappin.New(
					wappin.WithHystrixOptions(hystrix.WithErrorPercentThreshold(100)),
					wappin.WithClient(ts.doer),
				)
			},
			expect: func() (*wappin.ResponseMessage, error) {
				return &wappin.ResponseMessage{
						Status:         "401",
						Message:        "error invalid token",
						HttpStatusCode: http.StatusUnauthorized,
						RawData:        `{"status":"401","message":"error invalid token"}`,
					}, &wappin.Error{
						Status:  "401",
						Message: "error invalid token",
					}
			},
			expectErr:              false,
			sendMessageCalledTimes: 2,
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
				assert.Equal(t, tc.sendMessageCalledTimes, ts.doer.SendMessageCalledTimes)

				return
			}

			if err != nil {
				assert.Equal(t, expectErr.Error(), err.Error())
			}
			assert.Equal(t, expectResp, response)
			assert.Equal(t, tc.sendMessageCalledTimes, ts.doer.SendMessageCalledTimes)
		})
	}

}
