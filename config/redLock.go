package config

type RedLock struct {
	Hosts  []string `yaml:"hosts"`
	Passes []string `yaml:"passes"`
}
