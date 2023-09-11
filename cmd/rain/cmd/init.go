package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	initCmd = &cobra.Command{
		Use:     "init",
		Aliases: []string{"initialize", "initialise", "create"},
		Short:   "Initialize a Cobra Application",
		Run: func(_ *cobra.Command, _ []string) {
			projectPath, err := initializeProject()
			cobra.CheckErr(err)
			cobra.CheckErr(goGet("github.com/yrbb/rain"))
			cobra.CheckErr(goGet("github.com/spf13/cobra"))
			cobra.CheckErr(goTidy())
			fmt.Printf("Your Rain application is ready at\n%s\n", projectPath)
		},
	}
)

func initializeProject() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	modName := getModImportPath()

	project := &Project{
		AbsolutePath: wd,
		PkgName:      modName,
		AppName:      path.Base(modName),
	}

	if err := project.Create(); err != nil {
		return "", err
	}

	return project.AbsolutePath, nil
}

func getModImportPath() string {
	mod, cd := parseModInfo()
	return path.Join(mod.Path, fileToURL(strings.TrimPrefix(cd.Dir, mod.Dir)))
}

func fileToURL(in string) string {
	i := strings.Split(in, string(filepath.Separator))
	return path.Join(i...)
}

func parseModInfo() (Mod, CurDir) {
	var mod Mod
	var dir CurDir

	m := modInfoJSON("-m")
	cobra.CheckErr(json.Unmarshal(m, &mod))

	// Unsure why, but if no module is present Path is set to this string.
	if mod.Path == "command-line-arguments" {
		cobra.CheckErr("Please run `go mod init <MODNAME>` before `cobra-cli init`")
	}

	e := modInfoJSON("-e")
	cobra.CheckErr(json.Unmarshal(e, &dir))

	return mod, dir
}

type Mod struct {
	Path, Dir, GoMod string
}

type CurDir struct {
	Dir string
}

func goGet(mod string) error {
	return exec.Command("go", "get", mod).Run()
}

func goTidy() error {
	return exec.Command("go", "mod", "tidy").Run()
}

func modInfoJSON(args ...string) []byte {
	cmdArgs := append([]string{"list", "-json"}, args...)
	out, err := exec.Command("go", cmdArgs...).Output()
	cobra.CheckErr(err)

	return out
}
