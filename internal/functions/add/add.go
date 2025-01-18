package add

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/protomoks/pmok/internal/config"
	"github.com/protomoks/pmok/internal/utils"
)

type AddFunctionCommand struct {
	Name           string
	HttpPath       string
	AllowedMethods []string
}

var (
	//go:embed templates/function_template.ts
	functionTemplateEmbed []byte
)

func (a AddFunctionCommand) Valid() error {
	if a.Name == "" {
		return errors.New("name is required")
	}
	if a.HttpPath == "" {
		return errors.New("path is required")
	}
	return nil
}

func AddFunction(c AddFunctionCommand) error {
	if err := c.Valid(); err != nil {
		return err
	}
	// get the config
	conf := config.GetConfig()
	if conf == nil {
		return utils.ConfigNotFound()
	}
	// don't allow potentially overwriting existing functions
	if _, ok := conf.Manifest.Functions[c.Name]; ok {
		return fmt.Errorf("function with name %s already exists", c.Name)
	}

	if err := createFunctionFile(conf.GetProjectDir(), c.Name); err != nil {
		return err
	}

	conf.Manifest.Functions[c.Name] = config.Function{
		HttpPathname:   c.HttpPath,
		AllowedMethods: c.AllowedMethods,
		Entrypoint:     "index.ts",
	}

	return conf.Commit()
}

func createFunctionFile(projectRoot, name string) error {
	dir := filepath.Join(projectRoot, config.FunctionsDir, name)
	if err := os.Mkdir(dir, 0755); err != nil {
		return err
	}
	file, err := os.Create(filepath.Join(dir, "index.ts"))
	if err != nil {
		return err
	}
	defer file.Close()
	buf := bytes.NewBuffer(functionTemplateEmbed)
	_, err = io.Copy(file, buf)
	return err
}
