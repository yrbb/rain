package cmd

import (
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/yrbb/rain/cmd/rain/tpl"
)

type Project struct {
	PkgName      string
	AbsolutePath string
	AppName      string
}

func (p *Project) Create() error {
	if _, err := os.Stat(p.AbsolutePath); os.IsNotExist(err) {
		if err := os.Mkdir(p.AbsolutePath, 0754); err != nil {
			return err
		}
	}

	// create main.go
	mainFile, err := os.Create(fmt.Sprintf("%s/main.go", p.AbsolutePath))
	if err != nil {
		return err
	}
	defer mainFile.Close()

	mainTemplate := template.Must(template.New("main").Parse(string(tpl.MainTemplate())))
	err = mainTemplate.Execute(mainFile, p)
	if err != nil {
		return err
	}

	// create makefile
	makeFile, err := os.Create(fmt.Sprintf("%s/makefile", p.AbsolutePath))
	if err != nil {
		return err
	}
	makeFile.Write(tpl.MakefileTemplate())
	makeFile.Close()

	// create gitignore
	ignoreFile, err := os.Create(fmt.Sprintf("%s/.gitignore", p.AbsolutePath))
	if err != nil {
		return err
	}
	ignoreFile.Write(tpl.GitIgnoreTemplate())
	ignoreFile.Close()

	// create config/config.toml
	if _, err = os.Stat(fmt.Sprintf("%s/config", p.AbsolutePath)); os.IsNotExist(err) {
		cobra.CheckErr(os.Mkdir(fmt.Sprintf("%s/config", p.AbsolutePath), 0751))
	}
	configFile, err := os.Create(fmt.Sprintf("%s/config/config.toml", p.AbsolutePath))
	if err != nil {
		return err
	}
	configFile.Write(tpl.ConfigTemplate())
	configFile.Close()

	// create cmd/root.go
	if _, err = os.Stat(fmt.Sprintf("%s/cmd", p.AbsolutePath)); os.IsNotExist(err) {
		cobra.CheckErr(os.Mkdir(fmt.Sprintf("%s/cmd", p.AbsolutePath), 0751))
	}
	rootFile, err := os.Create(fmt.Sprintf("%s/cmd/root.go", p.AbsolutePath))
	if err != nil {
		return err
	}
	defer rootFile.Close()

	return template.Must(template.New("root").Parse(string(tpl.RootTemplate()))).Execute(rootFile, p)
}
