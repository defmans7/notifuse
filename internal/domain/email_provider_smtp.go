package domain

// SMTPWebhookPayload represents an SMTP webhook payload
// SMTP doesn't typically have a built-in webhook system, so this is a generic structure
// that could be used with a third-party SMTP provider that offers webhooks
type SMTPWebhookPayload struct {
	Event          string            `json:"event"`
	Timestamp      string            `json:"timestamp"`
	MessageID      string            `json:"message_id"`
	Recipient      string            `json:"recipient"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	Tags           []string          `json:"tags,omitempty"`
	Reason         string            `json:"reason,omitempty"`
	Description    string            `json:"description,omitempty"`
	BounceCategory string            `json:"bounce_category,omitempty"`
	DiagnosticCode string            `json:"diagnostic_code,omitempty"`
	ComplaintType  string            `json:"complaint_type,omitempty"`
}
