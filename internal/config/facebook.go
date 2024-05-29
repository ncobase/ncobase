package config

// Facebook github config struct
type Facebook struct {
	ID     string
	Secret string
}

func getFacebook() *Facebook {
	return &Facebook{
		ID:     c.GetString("facebook.id"),
		Secret: c.GetString("facebook.secret"),
	}
}
