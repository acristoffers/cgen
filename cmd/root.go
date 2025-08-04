package cmd

import (
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
		version, err := cmd.Flags().GetBool("version")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not parse options: %s\n", err)
			os.Exit(1)
		}

		if version {
			fmt.Printf("cgen version %s", cgen.Version)
			return
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
