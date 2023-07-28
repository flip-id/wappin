package v2

// RequestMessage is a request message for sending WhatsApp message.
// Type is required, the valid value is text, image, video, template and interactive
type RequestMessage struct {
	To            string          `json:"to"`
	Type          string          `json:"type"`
	RecipientType string          `json:"recipient_type"`
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
	Type       string                      `json:"type"`
	SubType    string                      `json:"subType"`
	Parameters []ComponentParameterRequest `json:"parameters"`
	Index      int                         `json:"index"`
}

// ComponentParameterRequest is required if the template has a dynamic variable value
type ComponentParameterRequest struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
