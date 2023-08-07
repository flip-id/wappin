package v2

// RequestMessage is a request message for sending WhatsApp message.
// Type is required, the valid value is text, image, video, template and interactive
type RequestMessage struct {
	To            string          `json:"to"`
	Type          string          `json:"type"`
	RecipientType string          `json:"recipient_type,omitempty"`
	Template      TemplateRequest `json:"template"`
}

// TemplateRequest is required to sending Whatsapp message with template format
type TemplateRequest struct {
	Name       string             `json:"name"`
	Language   LanguageRequest    `json:"language"`
	Namespace  string             `json:"namespace"`
	Components []ComponentRequest `json:"components"`
}

// LanguageRequest is required to sending Whatsapp with template format
// Policy is required, for default value is deterministic
type LanguageRequest struct {
	Policy string `json:"policy"`
	Code   string `json:"code"`
}

// ComponentRequest is required if the template has a dynamic variable value
// Type is contains valid value header, body and button
// SubType is optional, only used for button template
// Index is optional, used to set button position, only valid for templates with buttons
type ComponentRequest struct {
	Type       string                      `json:"type,omitempty"`
	SubType    string                      `json:"sub_type,omitempty"`
	Parameters []ComponentParameterRequest `json:"parameters,omitempty"`
	Index      *int                        `json:"index,omitempty"`
}

// ComponentParameterRequest is required if the template has a dynamic variable value
type ComponentParameterRequest struct {
	Type  string                 `json:"type"`
	Text  string                 `json:"text,omitempty"`
	Image *MediaParameterRequest `json:"image,omitempty"`
	Audio *MediaParameterRequest `json:"audio,omitempty"`
	Video *MediaParameterRequest `json:"video,omitempty"`
}

// MediaParameterRequest is required for Media request, select one Id or Link
// For Caption and FileName is optional
type MediaParameterRequest struct {
	Id       string `json:"id,omitempty"`
	Link     string `json:"link,omitempty"`
	Caption  string `json:"caption,omitempty"`
	FileName string `json:"file_name,omitempty"`
}
