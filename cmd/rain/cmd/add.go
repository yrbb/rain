package cmd

import (
	"fmt"
	"os"
	"unicode"

	"github.com/spf13/cobra"
)

var (
	parentName string

	addCmd = &cobra.Command{
		Use:     "add [command name]",
		Aliases: []string{"command"},
		Short:   "Add a command to a Rain Application",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				cobra.CheckErr(fmt.Errorf("add needs a name for the command"))
			}

			wd, err := os.Getwd()
			cobra.CheckErr(err)

			commandName := validateCmdName(args[0])
			command := &Command{
				CmdArgName: args[0],
				CmdName:    commandName,
				CmdParent:  parentName,
				Project: &Project{
					PkgName:      getModImportPath(),
					AbsolutePath: wd,
				},
			}

			cobra.CheckErr(command.Create())

			fmt.Printf("%s created at %s\n", command.CmdName, command.AbsolutePath)
		},
	}
)

func init() {
	addCmd.Flags().StringVarP(&parentName, "parent", "p", "rootCmd", "variable name of parent command for this command")
}

// validateCmdName returns source without any dashes and underscore.
// If there will be dash or underscore, next letter will be uppered.
// It supports only ASCII (1-byte character) strings.
// https://github.com/spf13/cobra/issues/269
func validateCmdName(source string) string {
	i := 0
	l := len(source)
	// The output is initialized on demand, then first dash or underscore
	// occurs.
	var output string

	for i < l {
		if source[i] == '-' || source[i] == '_' {
			if output == "" {
				output = source[:i]
			}

			// If it's last rune and it's dash or underscore,
			// don't add it output and break the loop.
			if i == l-1 {
				break
			}

			// If next character is dash or underscore,
			// just skip the current character.
			if source[i+1] == '-' || source[i+1] == '_' {
				i++
				continue
			}

			// If the current character is dash or underscore,
			// upper next letter and add to output.
			output += string(unicode.ToUpper(rune(source[i+1])))
			// We know, what source[i] is dash or underscore and source[i+1] is
			// uppered character, so make i = i+2.
			i += 2
			continue
		}

		// If the current character isn't dash or underscore,
		// just add it.
		if output != "" {
			output += string(source[i])
		}
		i++
	}

	if output == "" {
		return source // source is initially valid name.
	}
	return output
}
