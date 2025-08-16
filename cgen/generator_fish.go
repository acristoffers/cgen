package cgen

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"al.essio.dev/pkg/shellescape"
)

func GenerateFishCompletions(cli *CLI) error {
	dir := filepath.Join("share", "fish", "completions")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("could not create directory: %w", err)
	}

	path := filepath.Join(dir, fmt.Sprintf("%s.fish", cli.Name))
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer file.Close()

	return writeFishCompletions(cli, file)
}

func writeFishCompletions(cli *CLI, w io.Writer) error {
	for _, arg := range cli.Arguments {
		if _, err := fmt.Fprint(w, formatArgumentFish(cli.Name, arg, "")); err != nil {
			return err
		}
	}

	for _, cmd := range cli.Commands {
		if _, err := fmt.Fprint(w, formatCommandFish(cli.Name, cmd)); err != nil {
			return err
		}
	}

	return nil
}

func formatArgumentFish(cliName string, arg Argument, condition string) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("complete -c %s", cliName))
	if condition != "" {
		b.WriteString(fmt.Sprintf(" -n '%s'", condition))
	}

	if arg.Named {
		if arg.LongValueSeparator == "space" && arg.Completion.Type != "none" {
			b.WriteString(" -r")
		}

		if arg.SingleDashLong {
			if arg.Name != "" {
				b.WriteString(fmt.Sprintf(" -o %s", arg.Name))
			} else if arg.ShortName != "" {
				b.WriteString(fmt.Sprintf(" -o %s", arg.ShortName))
			}
		} else {
			if arg.ShortName != "" {
				b.WriteString(fmt.Sprintf(" -s %s", arg.ShortName))
			}
			if arg.Name != "" {
				b.WriteString(fmt.Sprintf(" -l %s", arg.Name))
			}
		}
	} else {
		b.WriteString(" -x")
	}

	desc := arg.ShortDescription
	if desc == "" {
		desc = arg.LongDescription
	}
	if desc == "" && !arg.Named {
		desc = arg.Name
	}
	if desc != "" {
		b.WriteString(fmt.Sprintf(" -d %s", shellescape.Quote(desc)))
	}

	switch arg.Completion.Type {
	case "file", "folder":
		b.WriteString(" -F")
	case "static":
		values := shellescape.Quote(strings.Join(arg.Completion.Values, " "))
		b.WriteString(fmt.Sprintf(" -fa %s", values))
	case "function":
		b.WriteString(fmt.Sprintf(" -fa %s", shellescape.Quote("("+arg.Completion.Fish+")")))
	}

	b.WriteString("\n")
	return b.String()
}

func formatCommandFish(cliName string, cmd Command) string {
	var b strings.Builder

	names := append([]string{cmd.Name}, cmd.Aliases...)
	for _, name := range names {
		b.WriteString(fmt.Sprintf("complete -c %s -n '__fish_use_subcommand' -f -a %s", cliName, name))
		if cmd.LongDescription != "" {
			b.WriteString(fmt.Sprintf(" -d %s", shellescape.Quote(cmd.LongDescription)))
		}
		b.WriteString("\n")

		for _, arg := range cmd.Arguments {
			b.WriteString(formatArgumentFish(cliName, arg, fmt.Sprintf("__fish_seen_subcommand_from %s", name)))
		}

		for _, subcmd := range cmd.Subcommands {
			b.WriteString(formatSubcommandFish(cliName, []string{name}, cmd, subcmd))
		}
	}

	return b.String()
}

func formatSubcommandFish(cliName string, path []string, parent Command, subcmd Command) string {
	var b strings.Builder

	names := append([]string{subcmd.Name}, subcmd.Aliases...)
	for _, name := range names {
		b.WriteString(fmt.Sprintf("complete -c %s", cliName))

		// Construct -n conditions
		andSeen := make([]string, 0, len(path))
		for _, p := range path {
			andSeen = append(andSeen, fmt.Sprintf("__fish_seen_subcommand_from %s", p))
		}
		andNot := []string{}
		for _, sibling := range parent.Subcommands {
			andNot = append(andNot, fmt.Sprintf("__fish_seen_subcommand_from %s", sibling.Name))
			for _, alias := range sibling.Aliases {
				andNot = append(andNot, fmt.Sprintf("__fish_seen_subcommand_from %s", alias))
			}
		}

		condition := strings.Join(andSeen, "; and ")
		if len(andNot) > 0 {
			condition += "; and not " + strings.Join(andNot, "; and not ")
		}
		if condition != "" {
			b.WriteString(fmt.Sprintf(" -n '%s'", condition))
		}

		b.WriteString(fmt.Sprintf(" -fa %s", name))
		if subcmd.LongDescription != "" {
			b.WriteString(fmt.Sprintf(" -d %s", shellescape.Quote(subcmd.LongDescription)))
		}
		b.WriteString("\n")

		andSeen = append(andSeen, fmt.Sprintf("__fish_seen_subcommand_from %s", subcmd.Name))
		for _, alias := range subcmd.Aliases {
			andSeen = append(andSeen, fmt.Sprintf("__fish_seen_subcommand_from %s", alias))
		}
		for _, arg := range subcmd.Arguments {
			b.WriteString(formatArgumentFish(cliName, arg, strings.Join(andSeen, "; and ")))
		}

		for _, nested := range subcmd.Subcommands {
			b.WriteString(formatSubcommandFish(cliName, append(path, name), subcmd, nested))
		}
	}

	return b.String()
}
