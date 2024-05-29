package config

// Github github config struct
type Github struct {
	ID     string
	Secret string
}

func getGithub() *Github {
	return &Github{
		ID:     c.GetString("github.id"),
		Secret: c.GetString("github.secret"),
	}
}
