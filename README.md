# cgen

Generates completion scripts for Fish, Bash and ZSH from a description file.

## Description file

The file that describes the CLI tool with all its commands, subcommands and arguments has the
following format (note that this is an example, and not meant to be a real completion suit for git):

```yaml
name: "git"
short-description: "Version control system"
long-description: "Git is a fast, scalable, distributed revision control system."
version: "2.42.0"

arguments: # global arguments
  - named: true # false for positional arguments, true for named ones (like -v, --verbose, -name)
    name: "version"
    short-description: "Print the git version" # shown during completion by some shells
    single-dash-long: false # defaults to false, no need to pass it every time
    space-value-separator: false # defaults to true
    equal-value-separator: false # defaults to false
    chainable: true # defaults to true
    completion: # this is the default
      type: "none" # none -> no completion
                   # file -> file completion
                   # folder -> folder completion
                   # static -> a static list of values (see repository below)
                   # function -> use the output of a command (see branch below)
    long-description: "Show git’s version and exit." # for man pages only
    example: "--version"

  - named: true
    name: "help"
    short-description: "Print help message"
    short-name: "h"
    single-dash-long: false # single-dash-long:true is like the find tool: -name is an option
                            # single-dash-long:false is like most GNU tools: --name is an option, -name == -n -a -m -e
    chainable: true # if the short-version can be part of a chain, like "-hal" == "-h -a -l".
    completion:
      type: "none"
    example: "-h" # for man pages only

commands:
  - name: "clone"
    usage: "clone <repository> [directory]" # for man pages only
    long-description: "Clone a repository into a new directory."
    arguments:
      - named: false
        name: "repository"
        short-description: "Repository URL"
        completion:
          type: "static"
          values: ["https://github.com/user/repo.git"]
        example: "git clone https://github.com/user/repo.git"

      - named: false
        name: "directory"
        short-description: "Target directory"
        completion:
          type: "folder"
        example: "git clone https://... my-folder"

  - name: "commit"
    usage: "commit -m <message>"
    arguments:
      - named: true
        name: "message"
        short-name: "m"
        short-description: "Commit message"
        space-value-separator: true
        equal-value-separator: true
        chainable: true
        completion:
          type: "none"
        example: "git commit -m 'Initial commit'"

  - name: "push"
    usage: "push [remote] [branch]"
    arguments:
      - named: false
        name: "remote"
        short-description: "Remote name"
        completion:
          type: "static"
          values: ["origin", "upstream"] # those two values will be offered as completion
        example: "git push origin"

      - named: false
        name: "branch"
        short-description: "Branch name"
        completion:
          type: "function" # the function is executed inside $(), so you can have pipes | here
          bash: "git branch --format='%(refname:short)'"
          fish: "git branch --format='%(refname:short)'"
          zsh: "git branch --format='%(refname:short)'"
        example: "git push origin main"

  - name: "remote"
    commands: # sub-commands
      - name: "add"
        usage: "remote add <name> <url>"
        arguments:
          - named: false
            name: "name"
            completion:
              type: "none"
            example: "git remote add upstream"

          - named: false
            name: "url"
            completion:
              type: "none"
            example: "git remote add upstream https://github.com/..."
```

### State

Not everything is implemented yet (or even possible in all generators).

What is missing:

- Man pages
- arguments:chainable (does nothing/ignored. Do completion engines even support this?)
- equal-value-separator and space-value-separator (used sometimes, but not really enforced)
- A mechanism to not offer some completion if another option is already in the command line.
- completion:type="folder" is the same as "file" for now

I did not really test positional arguments. They are implemented though.

What is known to work:

- Command completion.
- Positional arguments (In ZSH, a command that accepts subcommands cannot have positional arguments,
  they will be ignored. In BASH and Fish completion for subcommands and positional arguments are
  joined).
- Named argument completion.
- Named argument's value (all completion:type have been tested, and with the exception that "folder"
  and "file" do the same thing, they work)
- Subcommands (command nesting).
