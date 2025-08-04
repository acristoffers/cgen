package generators

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/acristoffers/cgen/cgen"
	"al.essio.dev/pkg/shellescape"
)

func GenerateBashCompletions(cli *cgen.CLI) error {
	dir := filepath.Join("share", "bash", "completions")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("could not create directory: %w", err)
	}

	path := filepath.Join(dir, fmt.Sprintf("%s.bash", cli.Name))
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer file.Close()

	return writeBashCompletions(cli, file)
}

func writeBashCompletions(cli *cgen.CLI, w io.Writer) error {
	iw := newIndentedWriter(w, "  ")

	// Write helper function
	iw.WriteLine(`__bash_seen_word() {
  local word
  for word in "${COMP_WORDS[@]}"; do
    [[ "$word" == "$1" ]] && return 0
  done
  return 1
}

`)

	// _complete_command() stub
	iw.WriteLine("_complete_command() {\n")
	iw.Indent(func() error {
		iw.WriteLine("prev=${COMP_WORDS[COMP_CWORD-1]}\n")
		for _, cmd := range cli.Commands {
			if err := completeCommandBash(iw, cli, &cmd); err != nil {
				return err
			}
		}
		return nil
	})
	iw.WriteLine("}\n\n")

	// main CLI function
	iw.WriteLine(fmt.Sprintf("_%s() {\n", cli.Name))
	iw.Indent(func() error {
		writeBashArray(iw, "global_options", collectArguments(cli.Arguments))
		writeBashArray(iw, "commands", collectCommands(cli.Commands))
		iw.WriteLine("for command in \"${commands[@]}\"; do\n")
		iw.Indent(func() error {
			iw.WriteLine("if __bash_seen_word \"$command\"; then\n")
			iw.Indent(func() error {
				iw.WriteLine("_complete_command \"$command\"\n")
				iw.WriteLine("return $?\n")
				return nil
			})
			iw.WriteLine("fi\n")
			return nil
		})
		iw.WriteLine("done\n")
		iw.WriteLine("completions=(\"${commands[@]}\" \"${global_options[@]}\")\n")
		iw.WriteLine("COMPREPLY=( $(compgen -W \"${completions[*]}\" -- \"${COMP_WORDS[COMP_CWORD]}\") )\n")
		return nil
	})
	iw.WriteLine(fmt.Sprintf("}\n\ncomplete -o bashdefault -F _%s %s\n", cli.Name, cli.Name))
	return nil
}

func writeBashArray(w *indentedWriter, name string, values []string) {
	sort.Strings(values)
	escaped := make([]string, len(values))
	for i, v := range values {
		escaped[i] = shellescape.Quote(v)
	}
	w.WriteLine(fmt.Sprintf("%s=(%s)\n", name, strings.Join(escaped, " ")))
}

func collectArguments(args []cgen.Argument) []string {
	out := []string{}
	for _, arg := range args {
		if arg.Name != "" {
			if arg.SingleDashLong {
				out = append(out, "-"+arg.Name)
			} else {
				out = append(out, "--"+arg.Name)
			}
		}
		if arg.ShortName != "" {
			out = append(out, "-"+arg.ShortName)
		}
	}
	return out
}

func collectCommands(cmds []cgen.Command) []string {
	out := []string{}
	for _, cmd := range cmds {
		out = append(out, cmd.Name)
		out = append(out, cmd.Aliases...)
	}
	return out
}

func completeCommandBash(w *indentedWriter, cli *cgen.CLI, cmd *cgen.Command) error {
	cmdNames := append([]string{cmd.Name}, cmd.Aliases...)

	for _, name := range cmdNames {
		w.WriteLine(fmt.Sprintf("if __bash_seen_word %s; then\n", shellescape.Quote(name)))
		err := w.Indent(func() error {
			writeBashArray(w, "arguments", collectArguments(cmd.Arguments))
			writeBashArray(w, "subcommands", collectCommands(cmd.Subcommands))

			if err := writeArgumentCaseSwitch(w, cmd.Arguments); err != nil {
				return err
			}

			w.WriteLine("completions=(\"${subcommands[@]}\" \"${arguments[@]}\" \"${global_options[@]}\")\n")
			w.WriteLine("COMPREPLY=( $(compgen -W \"${completions[*]}\" -- \"${COMP_WORDS[COMP_CWORD]}\") )\n")

			for _, sub := range cmd.Subcommands {
				err := completeSubcommandBash(w, []string{name}, &sub, cli.Arguments)
				if err != nil {
					return err
				}
			}

			w.WriteLine("return\n")

			return nil
		})
		if err != nil {
			return err
		}
		w.WriteLine("fi\n")
	}
	return nil
}

func completeSubcommandBash(w *indentedWriter, path []string, subcmd *cgen.Command, globalArgs []cgen.Argument) error {
	names := append([]string{subcmd.Name}, subcmd.Aliases...)

	for _, name := range names {
		w.WriteLine(fmt.Sprintf("if __bash_seen_word %s; then\n", shellescape.Quote(name)))
		err := w.Indent(func() error {
			writeBashArray(w, "arguments", collectArguments(subcmd.Arguments))
			writeBashArray(w, "subcommands", collectCommands(subcmd.Subcommands))

			if err := writeArgumentCaseSwitch(w, subcmd.Arguments); err != nil {
				return err
			}

			w.WriteLine("completions=(\"${subcommands[@]}\" \"${arguments[@]}\" \"${global_options[@]}\")\n")
			w.WriteLine("COMPREPLY=( $(compgen -W \"${completions[*]}\" -- \"${COMP_WORDS[COMP_CWORD]}\") )\n")

			for _, sub := range subcmd.Subcommands {
				err := completeSubcommandBash(w, append(path, name), &sub, globalArgs)
				if err != nil {
					return err
				}
			}

			w.WriteLine("return\n")

			return nil
		})
		if err != nil {
			return err
		}
		w.WriteLine("fi\n")
	}
	return nil
}

func writeArgumentCaseSwitch(w *indentedWriter, args []cgen.Argument) error {
	w.WriteLine("case \"$prev\" in\n")
	for _, arg := range args {
		keys := []string{}
		if arg.ShortName != "" {
			keys = append(keys, "-"+arg.ShortName)
		}
		if arg.Name != "" {
			if arg.SingleDashLong {
				keys = append(keys, "-"+arg.Name)
			} else {
				keys = append(keys, "--"+arg.Name)
			}
		}
		if arg.Completion.Type != "none" && len(keys) > 0 {
			w.WriteLine(fmt.Sprintf("%s)\n", strings.Join(keys, "|")))
			err := w.Indent(func() error {
				switch arg.Completion.Type {
				case "static":
					escaped := shellescape.QuoteCommand(arg.Completion.Values)
					w.WriteLine(fmt.Sprintf("values=(%s)\n", escaped))
					w.WriteLine("COMPREPLY=( $(compgen -W \"${values[*]}\" -- \"${COMP_WORDS[COMP_CWORD]}\") )\n")
				case "file":
					w.WriteLine("COMPREPLY=( $(compgen -f -- \"${COMP_WORDS[COMP_CWORD]}\") )\n")
				case "folder":
					w.WriteLine("COMPREPLY=( $(compgen -d -- \"${COMP_WORDS[COMP_CWORD]}\") )\n")
				case "function":
					w.WriteLine(fmt.Sprintf("values=$(%s)\n", arg.Completion.Bash))
					w.WriteLine("COMPREPLY=( $(compgen -W \"${values[*]}\" -- \"${COMP_WORDS[COMP_CWORD]}\") )\n")
				}
				w.WriteLine("return\n")
				return nil
			})
			if err != nil {
				return err
			}
			w.WriteLine(";;\n")
		}
	}
	w.WriteLine("esac\n")
	return nil
}
