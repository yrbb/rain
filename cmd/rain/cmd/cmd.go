package cmd

import (
	"fmt"
	"os"
	"text/template"

	"github.com/yrbb/rain/cmd/rain/tpl"
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

	return template.Must(template.New("sub").Parse(string(tpl.AddCommandTemplate()))).Execute(cmdFile, c)
}
