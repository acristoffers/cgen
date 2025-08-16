package cgen

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"al.essio.dev/pkg/shellescape"
)

func GenerateBashCompletions(cli *CLI) error {
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

func writeBashCompletions(cli *CLI, w io.Writer) error {
	iw := newIndentedWriter(w, "  ")

	// Write helper function
	iw.WriteLine(`__bash_seen_word() {
  local word
  for word in "${COMP_WORDS[@]}"; do
    [[ "$word" == "$1" ]] && return 0
  done
  return 1
}

index_of() {
  local val="$1"; shift
  for i in "${!COMP_WORDS[@]}"; do
    if [[ "${COMP_WORDS[$i]}" == "$val" ]]; then
      echo "$i"
      return
    fi
  done
  echo "-1"
}

`)

	// _complete_command() stub
	iw.WriteLine(fmt.Sprintf("_complete_command_%s() {\n", cli.Name))
	iw.Indent(func() error {
		writeBashArray(iw, "global_options", collectArguments(cli.Arguments))
		iw.WriteLine("prev=${COMP_WORDS[$((COMP_CWORD-1))]}\n")
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
				iw.WriteLine(fmt.Sprintf("_complete_command_%s \"$command\"\n", cli.Name))
				iw.WriteLine("return $?\n")
				return nil
			})
			iw.WriteLine("fi\n")
			return nil
		})
		iw.WriteLine("done\n")

		iw.WriteLine("prev=''\n")
		iw.WriteLine("if [ $COMP_CWORD -gt 0 ]; then\n")
		iw.Indent(func() error {
			iw.WriteLine("prev=${COMP_WORDS[$((COMP_CWORD-1))]}\n")
			return nil
		})
		iw.WriteLine("fi\n")
		if err := writeArgumentCaseSwitch(iw, cli.Arguments); err != nil {
			return err
		}

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

func collectArguments(args []Argument) []string {
	out := []string{}
	for _, arg := range args {
		if arg.Named {
			if arg.Name != "" {
				name := ""
				if arg.SingleDashLong {
					name = "-" + arg.Name
				} else {
					name = "--" + arg.Name
				}
				switch arg.LongValueSeparator {
				case "space":
					out = append(out, name)
				case "equal":
					out = append(out, name+"=")
				case "both":
					out = append(out, name)
					out = append(out, name+"=")
				default:
					log.Fatalf("The option %s is not valid for long-value-separator. Accepted values are space, equal and both", arg.LongValueSeparator)
					os.Exit(1)
				}
			}
			if arg.ShortName != "" {
				out = append(out, "-"+arg.ShortName)
			}
		} else {
			switch arg.Completion.Type {
			case "static":
				for _, value := range arg.Completion.Values {
					out = append(out, shellescape.Quote(value))
				}
			case "function":
				out = append(out, fmt.Sprintf("$(%s)", arg.Completion.Bash))
			}
		}
	}
	return out
}

func collectCommands(cmds []Command) []string {
	out := []string{}
	for _, cmd := range cmds {
		out = append(out, cmd.Name)
		out = append(out, cmd.Aliases...)
	}
	return out
}

func completeCommandBash(w *indentedWriter, cli *CLI, cmd *Command) error {
	cmdNames := append([]string{cmd.Name}, cmd.Aliases...)

	for _, name := range cmdNames {
		w.WriteLine(fmt.Sprintf("if __bash_seen_word %s; then\n", shellescape.Quote(name)))
		err := w.Indent(func() error {
			for _, sub := range cmd.Subcommands {
				err := completeSubcommandBash(w, []string{name}, &sub, cli.Arguments)
				if err != nil {
					return err
				}
			}

			writeBashArray(w, "arguments", collectArguments(cmd.Arguments))
			writeBashArray(w, "subcommands", collectCommands(cmd.Subcommands))
			w.WriteLine(fmt.Sprintf("pos=$((${#COMP_WORDS[@]} - $(index_of %s) - 1))\n", cmd.Name))

			if err := writeArgumentCaseSwitch(w, cmd.Arguments); err != nil {
				return err
			}

			w.WriteLine("completions=(\"${subcommands[@]}\" \"${arguments[@]}\" \"${global_options[@]}\")\n")
			w.WriteLine("COMPREPLY=( $(compgen -W \"${completions[*]}\" -- \"${COMP_WORDS[COMP_CWORD]}\") )\n")

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

func completeSubcommandBash(w *indentedWriter, path []string, subcmd *Command, globalArgs []Argument) error {
	names := append([]string{subcmd.Name}, subcmd.Aliases...)

	for _, name := range names {
		w.WriteLine(fmt.Sprintf("if __bash_seen_word %s; then\n", shellescape.Quote(name)))
		err := w.Indent(func() error {
			for _, sub := range subcmd.Subcommands {
				err := completeSubcommandBash(w, append(path, name), &sub, globalArgs)
				if err != nil {
					return err
				}
			}

			writeBashArray(w, "arguments", collectArguments(subcmd.Arguments))
			writeBashArray(w, "subcommands", collectCommands(subcmd.Subcommands))
			w.WriteLine(fmt.Sprintf("pos=$((${#COMP_WORDS[@]} - $(index_of %s) - 1))\n", subcmd.Name))

			if err := writeArgumentCaseSwitch(w, subcmd.Arguments); err != nil {
				return err
			}

			w.WriteLine("completions=(\"${subcommands[@]}\" \"${arguments[@]}\" \"${global_options[@]}\")\n")
			w.WriteLine("COMPREPLY=( $(compgen -W \"${completions[*]}\" -- \"${COMP_WORDS[COMP_CWORD]}\") )\n")

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

func writeArgumentCaseSwitch(w *indentedWriter, args []Argument) error {
	switch_prev_written := false
	w.WriteLine("cur=${COMP_WORDS[$COMP_CWORD]}\n")
	w.WriteLine("case \"$cur\" in\n")
	goto swith_body

switch_prev:
	switch_prev_written = true
	w.WriteLine("case \"$prev\" in\n")
swith_body:
	for _, arg := range args {
		if !arg.Named {
			continue
		}
		keys := []string{}
		if arg.ShortName != "" {
			keys = append(keys, "-"+arg.ShortName)
		}
		name := ""
		if arg.Name != "" {
			if arg.SingleDashLong {
				name = "-" + arg.Name
			} else {
				name = "--" + arg.Name
			}
			switch arg.LongValueSeparator {
			case "space":
				if switch_prev_written {
					keys = append(keys, name)
				}
			case "equal":
				if !switch_prev_written {
					keys = append(keys, name+"=")
				}
			case "both":
				if switch_prev_written {
					keys = append(keys, name)
				}
				if !switch_prev_written {
					keys = append(keys, name+"=")
				}
			default:
				log.Fatalf("The option %s is not valid for long-value-separator. Accepted values are space, equal and both", arg.LongValueSeparator)
				os.Exit(1)
			}
		}
		if arg.Completion.Type != "none" && len(keys) > 0 {
			for _, key := range keys {
				w.WriteLine(fmt.Sprintf("%s)\n", key))
				err := w.Indent(func() error {
					switch arg.Completion.Type {
					case "static":
						escaped := shellescape.QuoteCommand(arg.Completion.Values)
						if !switch_prev_written {
							xs := []string{}
							for _, value := range arg.Completion.Values {
								xs = append(xs, key+value)
							}
							escaped = strings.Join(xs, " ")
						}
						w.WriteLine(fmt.Sprintf("values=(%s)\n", escaped))
						w.WriteLine("COMPREPLY=( $(compgen -W \"${values[*]}\" -- \"${COMP_WORDS[COMP_CWORD]}\") )\n")
					case "file":
						w.WriteLine("COMPREPLY=( $(compgen -f -- \"${COMP_WORDS[COMP_CWORD]}\") )\n")
					case "folder":
						w.WriteLine("COMPREPLY=( $(compgen -d -- \"${COMP_WORDS[COMP_CWORD]}\") )\n")
					case "function":
						w.WriteLine(fmt.Sprintf("values=($(%s))\n", arg.Completion.Bash))
						if !switch_prev_written {
							w.WriteLine(fmt.Sprintf("values=(${values[@]/#/%s})\n", key))
						}
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
	}
	w.WriteLine("esac\n")
	if !switch_prev_written {
		goto switch_prev
	}

	w.WriteLine("case \"$pos\" in\n")
	w.Indent(func() error {
		count := 1
		for _, arg := range args {
			if arg.Named {
				continue
			}
			w.WriteLine(fmt.Sprintf("%d)\n", count))
			w.Indent(func() error {
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
				if count == 1 {
					w.WriteLine("COMPREPLY=(\"${COMPREPLY[@]}\" \"${subcommands[@]}\")\n")
				}
				w.WriteLine("return\n")
				return nil
			})
			w.WriteLine(";;\n")
			count++
		}
		return nil
	})
	w.WriteLine("esac\n")

	return nil
}
