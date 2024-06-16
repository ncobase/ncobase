package helper

import (
	"errors"
	"ncobase/common/email"
	"ncobase/common/log"

	"github.com/gin-gonic/gin"
)

// SetEmailSender sets email sender to gin.Context
func SetEmailSender(c *gin.Context, sender email.Sender) {
	SetValue(c, "email_sender", sender)
}

// GetEmailSender gets email sender from gin.Context based on the configured provider
func GetEmailSender(c *gin.Context) (email.Sender, error) {
	if sender, ok := GetValue(c, "email_sender").(email.Sender); ok {
		return sender, nil
	}

	// Get email config
	emailConfig := GetConfig(c).Email
	var emailProviderConfig email.Config

	// Determine which provider to use based on the configured provider
	switch emailConfig.Provider {
	case "mailgun":
		emailProviderConfig = &emailConfig.Mailgun
	case "aliyun":
		emailProviderConfig = &emailConfig.Aliyun
	case "netease":
		emailProviderConfig = &emailConfig.NetEase
	case "sendgrid":
		emailProviderConfig = &emailConfig.SendGrid
	case "smtp":
		emailProviderConfig = &emailConfig.SMTP
	case "tencent_cloud":
		emailProviderConfig = &emailConfig.TencentCloud
	default:
		return nil, errors.New("unknown email provider")
	}

	// Create email sender based on the configured provider
	sender, err := email.NewSender(emailProviderConfig)
	if err != nil {
		log.Errorf(c, "Error creating email sender: %v\n", err)
		return nil, err
	}

	// Set email sender to gin.Context for future use
	SetEmailSender(c, sender)
	return sender, nil
}

// SendEmailWithTemplate sends an email with a template
func SendEmailWithTemplate(c *gin.Context, recipientEmail string, template email.AuthEmailTemplate) (string, error) {
	sender, err := GetEmailSender(c)
	if err != nil {
		return "", err
	}
	return sender.SendTemplateEmail(recipientEmail, template)
}
