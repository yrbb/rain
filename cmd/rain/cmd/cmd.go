package cmd

import (
	"fmt"
	"os"
	"text/template"

	"github.com/yrbb/rain/cmd/rain/tpl"
	"github.com/yrbb/rain/pkg/utils"
)

type Command struct {
	CmdName        string
	CmdNameUcFirst string
	CmdPrefix      string
	CmdParent      string
	Advance        bool
	CmdArgName     string
	*Project
}

func (c *Command) Create() error {
	c.CmdPrefix = c.CmdName[:1]
	c.CmdNameUcFirst = c.CmdPrefix + c.CmdName[1:]

	cmdFile, err := os.Create(fmt.Sprintf("%s/cmd/%s.go", c.AbsolutePath, c.CmdArgName))
	if err != nil {
		return err
	}
	defer cmdFile.Close()

	if c.Advance {
		if err := template.Must(template.New("sub").
			Parse(string(tpl.AddCommandTemplateV2()))).
			Execute(cmdFile, c); err != nil {
			return err
		}

		cmdPkgPath := fmt.Sprintf("%s/pkg/%s", c.AbsolutePath, c.CmdArgName)
		if !utils.IsDir(cmdPkgPath) {
			if err := utils.MakeDir(cmdPkgPath); err != nil {
				return err
			}
		}

		pkgFile, err := os.Create(fmt.Sprintf("%s/%s.go", cmdPkgPath, c.CmdArgName))
		if err != nil {
			return err
		}
		defer pkgFile.Close()

		if err := template.Must(template.New("sub-pkg").
			Parse(string(tpl.AddCommandPkgTemplateV2()))).
			Execute(pkgFile, c); err != nil {
			return err
		}

		pkgConfigFile, err := os.Create(fmt.Sprintf("%s/config.go", cmdPkgPath))
		if err != nil {
			return err
		}
		defer pkgConfigFile.Close()

		return template.Must(template.New("sub-pkg-config").
			Parse(string(tpl.AddCommandPkgConfigTemplateV2()))).
			Execute(pkgConfigFile, c)
	}

	return template.Must(template.New("sub").Parse(string(tpl.AddCommandTemplate()))).Execute(cmdFile, c)
}
