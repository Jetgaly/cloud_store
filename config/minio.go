package config

type MinIO struct {
	Endpoint        string `yaml:"Endpoint"`
	AccessKeyID     string `yaml:"AccessKeyID"`
	SecretAccessKey string `yaml:"SecretAccessKey"`
	UseSSL          bool   `yaml:"UseSSL"`
	UploadBucket    string `yaml:"UploadBucket"`
}
