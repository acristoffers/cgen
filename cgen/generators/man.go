package generators

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/acristoffers/cgen/cgen"
)

func GenerateManPage(cli *cgen.CLI) error {
	dir := filepath.Join("share", "man", "man1")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("could not create directory: %w", err)
	}
	path := filepath.Join(dir, fmt.Sprintf("%s.1", cli.Name))
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer file.Close()
	if err := writeManPage(cli, nil, cli.Arguments, cli.Commands, []string{cli.Name}, file); err != nil {
		return err
	}
	for _, cmd := range cli.Commands {
		if err := generateManPageForCommand(cli, &cmd, append(cli.Arguments, cmd.Arguments...), cmd.Subcommands, []string{cli.Name, cmd.Name}); err != nil {
			return err
		}
	}
	return nil
}

func generateManPageForCommand(cli *cgen.CLI, cmd *cgen.Command, args []cgen.Argument, cmds []cgen.Command, parents []string) error {
	path := filepath.Join("share", "man", "man1", fmt.Sprintf("%s.1", strings.Join(parents, "-")))
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer file.Close()
	if err := writeManPage(cli, cmd, args, cmds, parents, file); err != nil {
		return err
	}
	if cmd != nil {
		for _, subcmd := range cmd.Subcommands {
			if err := generateManPageForCommand(cli, &subcmd, append(cli.Arguments, subcmd.Arguments...), subcmd.Subcommands, append(parents, subcmd.Name)); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeManPage(cli *cgen.CLI, cmd *cgen.Command, args []cgen.Argument, cmds []cgen.Command, parents []string, file io.Writer) error {
	fmt.Fprintf(file, ".TH %s 1 \"%s\" \"%s\"\n", strings.Join(parents, "-"), time.Now().Format("02-Jan-2006"), cli.Version)
	fmt.Fprint(file, ".SH NAME\n")
	if cmd == nil {
		fmt.Fprintf(file, "%s \\- %s\n", cli.Name, cli.ShortDescription)
	} else {
		fmt.Fprintf(file, "%s \\- %s\n", strings.Join(parents, "-"), cmd.ShortDescription)
	}
	fmt.Fprint(file, ".SH SYNOPSIS\n")
	if cmd == nil {
		fmt.Fprintf(file, ".B %s\n", strings.Join(parents, " "))
	} else {
		fmt.Fprintf(file, ".B %s\n", strings.Join(parents, "-"))
	}
	for _, arg := range formatManArguments(args) {
		fmt.Fprintf(file, "[%s] ", arg)
	}
	pos := formatManPositionalArguments(args)
	if len(cmds) > 0 {
		fmt.Fprint(file, "\\fI<command>\\fR\n")
	} else if len(pos) > 0 {
		for _, pos := range pos {
			fmt.Fprintf(file, "\\fI%s\\fR", strings.ToUpper(pos))
		}
	}
	fmt.Fprintln(file)
	fmt.Fprint(file, ".SH DESCRIPTION\n")
	if cmd == nil {
		fmt.Fprintf(file, ".B %s\n", cli.Name)
		fmt.Fprintf(file, "%s\n", cli.LongDescription)
	} else {
		fmt.Fprintf(file, ".B %s\n", strings.Join(parents, "-"))
		fmt.Fprintf(file, "%s\n", cmd.LongDescription)
		if cmd.Deprecated != "" {
			fmt.Fprint(file, ".B DEPRECATED\n")
			fmt.Fprintf(file, "%s\n", cmd.Deprecated)
		}
	}
	fmt.Fprint(file, ".SH OPTIONS\n")
	for _, arg := range args {
		fmt.Fprint(file, ".TP\n")
		fmt.Fprintf(file, "%s\n", formatManArgument(&arg, ", "))
		for _, line := range strings.Split(arg.LongDescription, "\n") {
			if line != "" {
				fmt.Fprintln(file, line)
			}
		}
		if arg.Deprecated != "" {
			fmt.Fprint(file, ".B DEPRECATED\n")
			fmt.Fprintf(file, "%s\n", arg.Deprecated)
		}
	}
	if len(cmds) > 0 {
		fmt.Fprint(file, ".SH COMMANDS\n")
		for _, cmd := range cmds {
			fmt.Fprint(file, ".TP\n")
			fmt.Fprintf(file, ".BR %s (1)\n", strings.Join(append(parents, cmd.Name), "-"))
			if cmd.Deprecated != "" {
				fmt.Fprint(file, ".B DEPRECATED\n")
				fmt.Fprintf(file, "%s\n", cmd.Deprecated)
			}
			if cmd.ShortDescription != "" {
				fmt.Fprintf(file, "%s\n", cmd.ShortDescription)
			}
		}
	}
	return nil
}

func formatManPositionalArguments(args []cgen.Argument) []string {
	xs := []string{}
	for _, arg := range args {
		if !arg.Named && !arg.Hidden {
			if arg.ValueLabel != "" {
				xs = append(xs, strings.ToUpper(arg.ValueLabel))
			} else if arg.Name != "" {
				xs = append(xs, strings.ToUpper(arg.Name))
			}
		}
	}
	return xs
}

func formatManArguments(args []cgen.Argument) []string {
	xs := []string{}
	for _, arg := range args {
		if !arg.Named || arg.Hidden {
			continue
		}
		xs = append(xs, formatManArgument(&arg, "|"))
	}
	return xs
}

func formatManArgument(arg *cgen.Argument, separator string) string {
	str := ""
	if arg.ShortName != "" {
		str = "\\fB\\-" + arg.ShortName + "\\fR"
		if arg.Completion.Type != "none" {
			switch arg.ShortValueSeparator {
			case "space", "both":
				str += " "
			case "attached":
			}
			str += "\\fI"
			if arg.ValueLabel != "" {
				str += strings.ToUpper(arg.ValueLabel)
			} else if arg.Name != "" {
				str += strings.ToUpper(arg.Name)
			} else {
				str += strings.ToUpper(arg.ShortName)
			}
			str += "\\fR"
		}
	}
	if arg.Name != "" {
		if str != "" {
			str += separator
		}
		str += "\\fB\\-"
		if !arg.SingleDashLong {
			str += "\\-"
		}
		str += arg.Name + "\\fR"
		if arg.Completion.Type != "none" {
			switch arg.LongValueSeparator {
			case "equal", "both":
				str += "="
			case "space":
				str += " "
			}
			str += "\\fI"
			if arg.ValueLabel != "" {
				str += strings.ToUpper(arg.ValueLabel)
			} else if arg.Name != "" {
				str += strings.ToUpper(arg.Name)
			} else {
				str += strings.ToUpper(arg.ShortName)
			}
			str += "\\fR"
		}
	}
	return str
}
