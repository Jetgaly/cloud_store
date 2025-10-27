package config

type Logger struct{
	Dev bool `yaml:"dev"`
	Level string `yaml:"level"` 
	OutputPaths []string `yaml:"outputPaths"`
	ErrorOutputPaths []string `yaml:"errorOutputPaths"`
}

