package config

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"
)

type Config struct {
	Manifest ManifestConfig
}

func (c *Config) GetProjectDir() string {
	return c.Manifest.rootDir
}

// Commits the contents of Config.Manifest to the manifest file
func (c Config) Commit() error {
	b, err := marshal(&c.Manifest)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(c.Manifest.ConfigPath(), os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, bytes.NewBuffer(b))
	return err
}

var (
	cfg  *Config
	once sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		wd, err := os.Getwd()
		if err != nil {
			return
		}
		dir, err := ResolveProjectDir(wd)
		if err != nil {
			return
		}
		mbytes, format, err := readManifestFile(dir)
		if err != nil {
			return
		}
		var manifest ManifestConfig
		if err := unmarshal(mbytes, &manifest, format); err != nil {
			return
		}
		manifest.format = format
		manifest.rootDir = dir
		cfg = &Config{
			Manifest: manifest,
		}
	})
	return cfg
}

func readManifestFile(dir string) ([]byte, ConfigFormat, error) {
	format := checkFormat(dir)
	manifestPath := DeploymentManifestYaml
	if format == ConfigJson {
		manifestPath = DeploymentManifestJson
	}
	mfile, err := os.Open(filepath.Join(dir, manifestPath))
	if err != nil {
		return nil, "", err
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, mfile); err != nil {
		return nil, format, err
	}
	return buf.Bytes(), format, nil
}

func checkFormat(loc string) ConfigFormat {
	// check if the json deployment.json exists. If not we probably have a yaml one
	if _, err := os.Stat(path.Join(loc, DeploymentManifestJson)); errors.Is(err, os.ErrNotExist) {
		return ConfigYaml
	}
	return ConfigJson
}
