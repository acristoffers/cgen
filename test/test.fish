#!/usr/bin/env fish

source share/fish/completions/cli.fish

complete -C "cli $argv" | cut -d\t -f1
