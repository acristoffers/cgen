package cmd

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/acristoffers/cgen/cgen"
	"github.com/acristoffers/cgen/cgen/generators"
	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "cgen [PATH]",
	Short: "Generates CLI completions from a configuration file",
	Long: `Generates Fish, BASH and ZSH completions for a tool from a yaml description file.

		This tool creates completion configuration files for Fish, BASH and ZSH based on a configuration
		file, allowing you to create completion files for existing tools.

		Usage:
			To generate the completion:
			- cgen config.yaml
			To generate an example configuration:
			- cgen --sample
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if version, err := cmd.Flags().GetBool("version"); err == nil && version {
			fmt.Printf("cgen version %s", cgen.Version)
			return
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "Could not parse options: %s\n", err)
			os.Exit(1)
		}

		if sample, err := cmd.Flags().GetBool("sample"); err == nil && sample {
			var buf bytes.Buffer
			enc := yaml.NewEncoder(&buf)
			enc.Encode(generateSample())
			os.Stdout.Write(buf.Bytes())
			return
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "Could not parse options: %s\n", err)
			os.Exit(1)
		}

		if len(args) != 1 {
			cmd.Help()
			os.Exit(0)
		}

		filePath, err := filepath.Abs(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not get configuration path: %s\n", err)
			os.Exit(1)
		}

		if _, err := os.Stat(filePath); err != nil {
			fmt.Fprintf(os.Stderr, "Could not open configuration file: %s\n", err)
			os.Exit(1)
		}

		binary, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not read configuration file: %s\n", err)
			os.Exit(1)
		}

		var cli cgen.CLI
		dec := yaml.NewDecoder(strings.NewReader(string(binary)), yaml.Validator(validator.New()), yaml.Strict())
		if err := dec.Decode(&cli); err != nil {
			fmt.Fprintf(os.Stderr, "Could not parse configuration file: %s\n", err)
			os.Exit(1)
		}

		if err := generators.GenerateBashCompletions(&cli); err != nil {
			log.Fatal("Error generating BASH completion: ", err.Error())
			os.Exit(1)
		}

		if err := generators.GenerateFishCompletions(&cli); err != nil {
			log.Fatal("Error generating Fish completion: ", err.Error())
			os.Exit(1)
		}

		if err := generators.GenerateZshCompletions(&cli); err != nil {
			log.Fatal("Error generating ZSH completion: ", err.Error())
			os.Exit(1)
		}

		if err := generators.GenerateManPage(&cli); err != nil {
			log.Fatal("Error generating man pages: ", err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.Flags().BoolP("version", "v", false, "Prints the version.")
	RootCmd.Flags().BoolP("sample", "s", false, "Prints a sample configuration.")
}

func generateSample() cgen.CLI {
	cli := cgen.CLI{}
	cli.Name = "cli"
	cli.Version = "0.0.1"
	cli.LongDescription = "cli long desc"
	cli.ShortDescription = "cli short desc"

	ga := cgen.Argument{}
	ga.Named = true
	ga.Name = "version"
	ga.ShortName = "v"
	ga.LongDescription = "version long desc"
	ga.ShortDescription = "version short desc"
	ga.Completion.Type = "none"
	ga.LongValueSeparator = "space"
	ga.ShortValueSeparator = "space"

	cli.Arguments = append(cli.Arguments, ga)

	cmd1 := cgen.Command{}
	cmd1.Name = "cmd1"
	cmd1.Aliases = []string{"c1"}
	cmd1.LongDescription = "cmd1 long description"
	cmd1.ShortDescription = "cmd1 short description"

	subcmd1 := cgen.Command{}
	subcmd1.Name = "subcmd1"
	subcmd1.Aliases = []string{"s1"}
	subcmd1.LongDescription = "subcmd1 long description"
	subcmd1.ShortDescription = "subcmd1 short description"
	subcmd1.Deprecated = "replaced by subcmd2"

	subcmd2 := cgen.Command{}
	subcmd2.Name = "subcmd2"
	subcmd2.Aliases = []string{"s2"}
	subcmd2.LongDescription = "subcmd2 long description"
	subcmd2.ShortDescription = "subcmd2 short description"

	cmd1.Subcommands = append(cmd1.Subcommands, subcmd1)
	cmd1.Subcommands = append(cmd1.Subcommands, subcmd2)

	cmd2 := cgen.Command{}
	cmd2.Name = "cmd2"
	cmd2.Aliases = []string{"c2"}
	cmd2.LongDescription = "cmd2 long description"
	cmd2.ShortDescription = "cmd2 short description"

	opt1 := cgen.Argument{}
	opt1.Named = true
	opt1.Name = "opt1"
	opt1.ShortName = "o"
	opt1.LongDescription = "opt1 long desc"
	opt1.ShortDescription = "opt1 short desc"
	opt1.LongValueSeparator = "equal"
	opt1.ShortValueSeparator = "attached"
	opt1.Completion.Type = "static"
	opt1.Completion.Values = []string{"a", "b", "c"}

	opt2 := cgen.Argument{}
	opt2.Named = true
	opt2.Name = "opt2"
	opt2.LongDescription = "opt2 long desc"
	opt2.ShortDescription = "opt2 short desc"
	opt2.Completion.Type = "function"
	opt2.Completion.Bash = "printf 'a\nb\nc'"
	opt2.Completion.Fish = "printf 'a\nb\nc'"
	opt2.Completion.Zsh = "printf 'a\nb\nc'"
	opt2.LongValueSeparator = "space"
	opt2.ShortValueSeparator = "space"
	opt2.Deprecated = "replaced by nothing"

	cmd2.Arguments = append(cmd2.Arguments, opt1)
	cmd2.Arguments = append(cmd2.Arguments, opt2)

	cli.Commands = append(cli.Commands, cmd1)
	cli.Commands = append(cli.Commands, cmd2)

	return cli
}
