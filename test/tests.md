# Global Named Arguments (ga)

1. complete named global argument
2. complete named global argument with "-"
3. complete value of named argument (static)
4. complete value of named argument (function)
5. complete named argument with =
6. complete value of named argument with = (static)
7. complete value of named argument with = (function)
8. complete value of short named argument (static)
9. complete value of short named argument (function)
10. complete value of short named argument without space (static)
11. complete value of short named argument without space (function)
12. complete argument after another argument

## Notes

- In case 1, fish has `~` in the output. The run script threats it as "ignore its output". This is
  so because fish will list files in this scenario, instead of the flags. That's why case 2 exists.
- In case 5, fish returns both `--verbose` and `--verbose=`. There is no way around it.

# Global Positional Arguments (gpa)

1. complete value of global positional argument (static)
2. complete value of global positional argument (function)

# Commands (cmd)

1. complete command name
2. complete command named argument
3. complete command named argument with "-"
4. complete value of command named argument (static)
5. complete value of command named argument (function)
6. complete value of command named argument with = (static)
7. complete value of command named argument with = (function)
8. complete value of command short named argument (static)
9. complete value of command short named argument (function)
10. complete value of command short named argument without space (static)
11. complete value of command short named argument without space (function)
12. complete global argument after command
13. complete positional argument after command

# Subcommand (sub)

1. complete subcommand name
2. complete subcommand named argument
3. complete value of subcommand named argument (static)
4. complete value of subcommand named argument (function)
5. complete value of subcommand named argument with = (static)
6. complete value of subcommand named argument with = (function)
7. complete value of subcommand short named argument (static)
8. complete value of subcommand short named argument (function)
9. complete value of subcommand short named argument without space (static)
10. complete value of subcommand short named argument without space (function)
11. complete global argument after subcommand
12. complete positional argument after subcommand
