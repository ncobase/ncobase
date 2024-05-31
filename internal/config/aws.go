package config

// AWS aws s3 config struct
type AWS struct {
	Region          string
	AccessKey       string
	AccessSecretKey string
	Bucket          string
}

func getAWSConfig() AWS {
	return AWS{
		Region:          c.GetString("aws.region"),
		AccessKey:       c.GetString("aws.access_key"),
		AccessSecretKey: c.GetString("aws.access_secret"),
		Bucket:          c.GetString("aws.bucket"),
	}
}
