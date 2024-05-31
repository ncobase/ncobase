package config

// Facebook facebook config struct
type Facebook struct {
	ID     string
	Secret string
}

func getFacebookConfig() Facebook {
	return Facebook{
		ID:     c.GetString("facebook.id"),
		Secret: c.GetString("facebook.secret"),
	}
}
