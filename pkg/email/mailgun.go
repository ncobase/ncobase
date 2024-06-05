package email

import (
	"context"
	"stocms/pkg/log"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/pkg/errors"
)

// MailgunConfig holds the configuration for Mailgun
type MailgunConfig struct {
	APIKey string
	Domain string
	From   string
}

// AuthEmailTemplate represents the email template for authentication
type AuthEmailTemplate struct {
	Subject  string `json:"subject"`
	Template string `json:"template"`
	Keyword  string `json:"keyword"`
	URL      string `json:"url"`
}

// SendTemplateEmailWithMailgun sends an email using Mailgun with the given template
func SendTemplateEmailWithMailgun(config *MailgunConfig, recipientEmail string, template AuthEmailTemplate) (string, error) {
	if err := validateMailgunConfig(config); err != nil {
		return "", err
	}

	mg := mailgun.NewMailgun(config.Domain, config.APIKey)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a new message with template
	message := mg.NewMessage(config.From, template.Subject, "")
	message.SetTemplate(template.Template)
	_ = message.AddRecipient(recipientEmail)
	message.AddVariable("keyword", template.Keyword)
	message.AddVariable("url", template.URL)

	// Send email
	_, id, err := mg.Send(ctx, message)
	if err != nil {
		log.Errorf(ctx, "Error sending email: %v", err)
		return "", err
	}

	log.Printf(ctx, "Email queued: %s", id)
	return id, nil
}

// validateMailgunConfig validates the Mailgun configuration
func validateMailgunConfig(config *MailgunConfig) error {
	if config.APIKey == "" || config.Domain == "" || config.From == "" {
		return errors.New("invalid Mailgun configuration")
	}
	return nil
}
