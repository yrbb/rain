package tpl

func MainTemplate() []byte {
	return []byte(`package main

import (
	"github.com/yrbb/rain"

	_ "{{ .PkgName }}/cmd"
)

func main() {
	app, err := rain.New()
	if err != nil {
		panic(err)
	}

	app.OnStart(func() {
		// do something
	})

	app.OnStop(func() {
		// do something
	})

	app.Run()
}
`)
}

func GitIgnoreTemplate() []byte {
	return []byte(`.git
.vscode
.idea
.DS_Store
bin`)
}

func ConfigTemplate() []byte {
	return []byte(`debug = true 
project = "rain"

[[redis]]
disable = true
name = "test-redis"
addr = "redis://password@127.0.0.1:6379"

[[database]]
disable = true
name = "test-mysql"
type = "mysql"
addr = "user:password@tcp(127.0.01:3306)/?charset=utf8mb4&interpolateParams=true"

[worker]
capacity = 1000

[custom]
test_key = "test_value"
`)
}

func MakefileTemplate() []byte {
	return []byte(`COMMIT_HASH=$(shell git rev-parse --short HEAD || echo "GitNotFound")
COMPILE=$(shell date '+%Y-%m-%d %H:%M:%S') by $(shell go version)
LDFLAGS="-X \"github.com/yrbb/rain.Version=${COMMIT_HASH}\" -X \"github.com/yrbb/rain.Compile=$(COMPILE)\""

.PHONY: all run build build-linux clean 
all: build

run: build
	@./bin/example test -c=config/config.toml

build: clean
	@go build -ldflags ${LDFLAGS} -o ./bin/example ./main.go

build-linux: clean
	@GOOS=linux GOARCH=amd64 go build -ldflags ${LDFLAGS} -o ./bin/example ./main.go

clean:
	@rm -rf bin
`)
}

func RootTemplate() []byte {
	return []byte(`package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yrbb/rain"
)

var rootCmd *cobra.Command = rain.RootCmd
`)
}

func AddCommandTemplate() []byte {
	return []byte(`package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var {{ .CmdName }}Cmd = &cobra.Command{
	Use:   "{{ .CmdName }}",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		{{ .CmdName }}Handler{}.Run(cmd, args)
	},
}

func init() {
	{{ .CmdParent }}.AddCommand({{ .CmdName }}Cmd)

	// {{ .CmdName }}Cmd.PersistentFlags().String("foo", "", "A help for foo")
	// {{ .CmdName }}Cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type {{ .CmdName }}Handler struct{}

func (g {{ .CmdName }}Handler) Run(cmd *cobra.Command, args []string) {
	fmt.Println("{{ .CmdName }} called")
}
`)
}
