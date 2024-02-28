package config

import (
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	Bind    string
	Debug   bool
	MaxInt  int
	MaxSize int
	Length  int
	Expire  int
	Dir     string
}

func (c *Config) GetDir() string {
	if c.Dir == "" {
		return filepath.Join(os.TempDir(), "gopload")
	}

	path, err := filepath.Abs(c.Dir)
	if err != nil {
		log.Panicf("path [%s] error: %s", c.Dir, err)
	}
	return path
}
