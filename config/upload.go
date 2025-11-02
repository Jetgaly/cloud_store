package config

type Upload struct {
	TempPath  string `yaml:"temp_path"`
	ChunkSize int64  `yaml:"chunk_size"`
}
