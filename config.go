package main

import (
	"os"
	"path/filepath"

	"github.com/phpgao/tlog"
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
		tlog.Fatalf("path [%s] error: %s", c.Dir, err)
	}
	tlog.Infof("dir: %s", path)
	return path
}
