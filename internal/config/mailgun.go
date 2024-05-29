package config

// Mailgun mailgun config struct
type Mailgun struct {
	Key    string
	Domain string
	From   string
}

func getMailgun() *Mailgun {
	return &Mailgun{
		Key:    c.GetString("mailgun.key"),
		Domain: c.GetString("mailgun.domain"),
		From:   c.GetString("mailgun.from"),
	}
}
