package v2

// BaseResponse is base response from Wappin consist meta version and error
type BaseResponse struct {
	Meta   MetaResponse `json:"meta"`
	Errors []Error      `json:"errors"`
}

// ResponseMessage is response from Wappin consist for success and error
type ResponseMessage struct {
	BaseResponse
	Messages []MessageResponse `json:"messages"`
}

// MetaResponse is version API from Wappin
type MetaResponse struct {
	Version string `json:"version"`
}

// MessageResponse is message ID from Wappin if you success sending Whatsapp message
type MessageResponse struct {
	Id string `json:"id"`
}

// Error response if you send invalid request to Wappin or something wrong issue from Wappin
type Error struct {
	Code    string `json:"code"`
	Title   string `json:"title"`
	Details string `json:"details"`
}

// ResponseLogin is response from login API Wappin to get credential
type ResponseLogin struct {
	BaseResponse
	User []UserResponse `json:"user"`
}

// UserResponse is to get Token and Expired time
type UserResponse struct {
	Token        string `json:"token"`
	ExpiredAfter string `json:"expired_after"`
}
