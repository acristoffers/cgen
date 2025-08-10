package generators

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"al.essio.dev/pkg/shellescape"
	"github.com/acristoffers/cgen/cgen"
)

func GenerateZshCompletions(cli *cgen.CLI) error {
	dir := filepath.Join("share", "zsh", "completions")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("could not create directory: %w", err)
	}

	path := filepath.Join(dir, fmt.Sprintf("_%s", cli.Name))
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer file.Close()

	return writeZshCompletions(cli, file)
}

func writeZshCompletions(cli *cgen.CLI, w io.Writer) error {
	iw := newIndentedWriter(w, "  ")

	iw.WriteLine(fmt.Sprintf("#compdef %s\n", cli.Name))
	iw.WriteLine(fmt.Sprintf("compdef _%s %s\n\n", cli.Name, cli.Name))

	for _, cmd := range cli.Commands {
		if err := writeZshFunctionCompletions(iw, &cmd); err != nil {
			return err
		}
	}

	for _, arg := range cli.Arguments {
		if arg.Completion.Type == "function" {
			iw.WriteLine(fmt.Sprintf("_arg_%s() {", arg.Name))
			iw.Indent(func() error {
				iw.WriteLine(fmt.Sprintf("compadd -- $(%s)", arg.Completion.Zsh))
				return nil
			})
			iw.WriteLine("}\n")
		}
	}

	iw.WriteLine(fmt.Sprintf("_%s() {\n", cli.Name))
	iw.Indent(func() error {
		cmds := collectCommands(cli.Commands)
		positionals := collectPositionalArgs(cli.Arguments)

		iw.WriteLine("_arguments -C \\\n")
		args := []string{}
		iw.Indent(func() error {
			count := 1
			if len(cmds) > 0 {
				args = append(args, fmt.Sprintf(`"%d:command:(%s)"`, count, strings.Join(cmds, " ")))
				count++
			}
			for _, arg := range positionals {
				line := formatZshPositional(count, arg)
				if line != "" {
					args = append(args, line)
					count++
				}
			}
			for _, arg := range cli.Arguments {
				if arg.Named {
					args = append(args, generateZshArgument(arg))
				}
			}
			if len(cmds) > 0 {
				args = append(args, "'*::args:->args'")
			}
			for i, arg := range args {
				iw.WriteLine(arg)
				if i < len(args)-1 {
					iw.WriteLine(" \\")
				}
				iw.WriteLine("\n")
			}
			return nil
		})

		iw.WriteLine("case $state in\n")
		iw.Indent(func() error {
			iw.WriteLine("args)\n")
			iw.Indent(func() error {
				iw.WriteLine("case ${words[1]} in\n")
				for _, cmd := range cli.Commands {
					writeZshCommandTree(iw, cli.Arguments, &cmd)
				}
				iw.WriteLine("esac\n")
				return nil
			})
			iw.WriteLine(";;\n")
			return nil
		})
		iw.WriteLine("esac\n")
		return nil
	})
	iw.WriteLine("}\n")
	return nil
}

func collectPositionalArgs(args []cgen.Argument) []cgen.Argument {
	var pos []cgen.Argument
	for _, arg := range args {
		if !arg.Named {
			pos = append(pos, arg)
		}
	}
	return pos
}

func formatZshPositional(count int, arg cgen.Argument) string {
	comp := ""
	switch arg.Completion.Type {
	case "none":
		return ""
	case "file", "folder":
		comp = "_files"
	case "static":
		comp = fmt.Sprintf("(%s)", shellescape.QuoteCommand(arg.Completion.Values))
	case "function":
		comp = fmt.Sprintf("_arg_%s", arg.Name)
	}
	return fmt.Sprintf(`"%d:%s:%s"`, count, arg.Name, comp)
}

func writeZshFunctionCompletions(w *indentedWriter, cmd *cgen.Command) error {
	for _, arg := range cmd.Arguments {
		if arg.Completion.Type == "function" {
			w.WriteLine(fmt.Sprintf("_arg_%s() {\n", arg.Name))
			w.Indent(func() error {
				w.WriteLine(fmt.Sprintf("compadd -- $(%s)\n", arg.Completion.Zsh))
				return nil
			})
			w.WriteLine("}\n\n")
		}
	}
	for _, sub := range cmd.Subcommands {
		if err := writeZshFunctionCompletions(w, &sub); err != nil {
			return err
		}
	}
	return nil
}

func generateZshArgument(arg cgen.Argument) string {
	comp := ""
	switch arg.Completion.Type {
	case "file", "folder":
		comp = "_files"
	case "static":
		comp = fmt.Sprintf("(%s)", shellescape.QuoteCommand(arg.Completion.Values))
	case "function":
		comp = fmt.Sprintf("_arg_%s", arg.Name)
	}
	dash := "-"
	if !arg.SingleDashLong {
		dash += "-"
	}
	desc := strings.ReplaceAll(arg.ShortDescription, "'", "'\"'\"'")
	longSeparatorSuffix := ""
	switch arg.LongValueSeparator {
	case "space":
	case "equal":
		longSeparatorSuffix = "=-"
	case "both":
		longSeparatorSuffix = "="
	default:
		log.Fatalf("The option %s is not valid for long-value-separator. Accepted values are space, equal and both", arg.LongValueSeparator)
		os.Exit(1)
	}
	name := arg.Name + longSeparatorSuffix
	shortSeparatorSuffix := ""
	switch arg.ShortValueSeparator {
	case "space":
	case "attached":
		shortSeparatorSuffix = "-"
	case "both":
		shortSeparatorSuffix = "+"
	default:
		log.Fatalf("The option %s is not valid for short-value-separator. Accepted values are space, equal and both", arg.ShortValueSeparator)
		os.Exit(1)
	}
	shortName := arg.ShortName + shortSeparatorSuffix
	if arg.Name != "" && arg.ShortName != "" {
		return fmt.Sprintf("'(-%s %s%s)'{-%s,%s%s}'[%s]:%s:%s'", shortName, dash, name, arg.ShortName, dash, name, desc, arg.Name, comp)
	} else if arg.Name != "" {
		return fmt.Sprintf("'%s%s[%s]:%s:%s'", dash, name, desc, arg.Name, comp)
	} else {
		return fmt.Sprintf("'-%s[%s]:%s:%s'", shortName, desc, arg.ShortName, comp)
	}
}

func writeZshCommandTree(w *indentedWriter, global_arguments []cgen.Argument, cmd *cgen.Command) error {
	if len(cmd.Arguments) == 0 && len(cmd.Subcommands) == 0 {
		return nil
	}

	names := append([]string{cmd.Name}, cmd.Aliases...)
	for _, name := range names {
		w.WriteLine(fmt.Sprintf("%s)\n", name))
		err := w.Indent(func() error {
			args := []string{}
			subs := collectCommands(cmd.Subcommands)
			if len(subs) > 0 {
				args = append(args, fmt.Sprintf("'1:command:(%s)'", strings.Join(subs, " ")))
			}
			if len(subs) == 0 {
				count := 1
				for _, arg := range cmd.Arguments {
					if !arg.Named {
						args = append(args, formatZshPositional(count, arg))
						count++
					}
				}
			}
			for _, arg := range cmd.Arguments {
				if arg.Named {
					args = append(args, generateZshArgument(arg))
				}
			}
			for _, arg := range global_arguments {
				if arg.Named {
					args = append(args, generateZshArgument(arg))
				}
			}
			if len(cmd.Subcommands) > 0 {
				args = append(args, fmt.Sprintf("'*::args:->args_%s'", cmd.Name))
			}
			if len(args) > 0 {
				w.WriteLine("_arguments -C \\\n")
				w.Indent(func() error {
					for i, line := range args {
						w.WriteLine(line)
						if i < len(args)-1 {
							w.WriteLine(" \\")
						}
						w.WriteLine("\n")
					}
					return nil
				})
			}
			if len(cmd.Subcommands) > 0 {
				w.WriteLine("case $state in\n")
				w.Indent(func() error {
					w.WriteLine(fmt.Sprintf("args_%s)\n", cmd.Name))
					w.Indent(func() error {
						w.WriteLine("case ${words[1]} in\n")
						w.Indent(func() error {
							for _, sub := range cmd.Subcommands {
								writeZshCommandTree(w, global_arguments, &sub)
							}
							return nil
						})
						w.WriteLine("esac\n")
						return nil
					})
					w.WriteLine(";;\n")
					return nil
				})
				w.WriteLine("esac\n")
			}
			return nil
		})
		if err != nil {
			return err
		}
		w.WriteLine(";;\n")
	}
	return nil
}
