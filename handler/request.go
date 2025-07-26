package handler

import "fmt"

func errParamIsRequired(name, typ string) error {
	return fmt.Errorf("param: %s (type: %s) is required", name, typ)
}

// CreateWebhook

type CreateWebhookRequest struct {
	URL string `json:"url"`
}

type SendMessageRequest struct {
	Message string `json:"message"`
	Sender  string `json:"sender"`
}

func (r *CreateWebhookRequest) Validate() error {
	if r.URL == "" {
		return errParamIsRequired("url", "string")
	}
	return nil
}

// UpdateWebhook

type UpdateWebhookRequest struct {
	Active *bool  `json:"active"`
	URL    string `json:"url"`
}

func (r *UpdateWebhookRequest) Validate() error {
	// If any field is provided, validation is truthy
	if r.Active != nil || r.URL != "" {
		return nil
	}
	// If none of the fields were provided, return falsy
	return fmt.Errorf("at least one valid field must be provided")
}
