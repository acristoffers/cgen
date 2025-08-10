# cgen

Generate shell completions for Fish, Bash, and ZSH from a single YAML file.

`cgen` lets you define your CLI interface once in YAML, and instantly generate shell completion
scripts for multiple shells. No more hand-writing separate completion files for every shell.

---

## ✨ Features

* **One source of truth**: describe your CLI in a single YAML file.
* **Multi-shell output**: generates completion scripts for **Fish**, **Bash**, and **ZSH**.
* **Commands & subcommands**: full nesting support.
* **Named & positional arguments**: completions work for both.
* **Static, file, folder, and dynamic completions**: choose from built-in types or run commands for values.
* **Shell-specific behavior**: define per-shell completion commands when needed.

---

## 📄 YAML Specification

Your YAML file defines:

1. **CLI metadata**
2. **Global arguments**
3. **Commands**
4. **Subcommands**
5. **Arguments (named & positional)**

Below is a breakdown of the supported fields.

---

### 1. CLI Metadata

```yaml
name: "git"
short-description: "Version control system"
long-description: "Git is a fast, scalable, distributed revision control system."
version: "2.42.0"
```

* `name` — the command name.
* `short-description` — shown in completion suggestions (when supported).
* `long-description` — optional extended description.
* `version` — your CLI tool version.

---

### 2. Global Arguments

These are available for **all** commands.

```yaml
arguments:
  - named: true
    name: "version"
    short-description: "Print the git version"
    single-dash-long: false
    long-value-separator: "space" # space | equal | both
    short-value-separator: "space" # space | attached | both
    completion:
      type: "none"
    long-description: "Show git’s version and exit."
    example: "--version"
```

**Fields:**

* `named`:

  * `true` → named option (`--verbose`, `-v`)
  * `false` → positional argument
* `name`: long form without leading dashes.
* `short-name`: optional short flag (`-v`).
* `single-dash-long`: whether long options can use a single dash (GNU-style vs. find-style).
* `long-value-separator`: how the long-option’s value is provided (`--opt value`, `--opt=value`, or both).
* `short-value-separator`: how the short-option’s value is provided (`-O 3`, `-O3`, or both).
* `completion`: completion behavior (see below).
* `example`: example usage for documentation/man pages.

---

### 3. Completion Types

```yaml
completion:
  type: "file"      # file completion
  type: "folder"    # folder completion
  type: "static"    # predefined list
  values: ["origin", "upstream"]

  type: "function"  # generated dynamically by running a command (executed inside "$()")
  bash: "git branch --format='%(refname:short)'"
  fish: "git branch --format='%(refname:short)'"
  zsh:  "git branch --format='%(refname:short)'"
```

**Supported types:**

* `none` — no completion.
* `file` — suggest files.
* `folder` — suggest folders (currently same as `file` in some shells).
* `static` — predefined values.
* `function` — run a shell command to generate suggestions.

---

### 4. Commands

```yaml
commands:
  - name: "clone"
    usage: "clone <repository> [directory]"
    long-description: "Clone a repository into a new directory."
    arguments:
      - named: false
        name: "repository"
        short-description: "Repository URL"
        completion:
          type: "static"
          values: ["https://github.com/user/repo.git"]

      - named: false
        name: "directory"
        short-description: "Target directory"
        completion:
          type: "folder"
```

Each command:

* Has a `name`.
* Can have its own `usage` and `long-description`.
* Can define its own arguments (named or positional).
* Can contain **subcommands** (nesting is supported).

---

### 5. Subcommands

```yaml
commands:
  - name: "remote"
    commands:
      - name: "add"
        usage: "remote add <name> <url>"
        arguments:
          - named: false
            name: "name"
            completion:
              type: "none"

          - named: false
            name: "url"
            completion:
              type: "none"
```

Subcommands are defined just like commands, but nested under a `commands` key.

---

## ✅ Currently Working

* Command & subcommand completion
* Named & positional arguments
* Static, file, folder, and dynamic completions
* Nesting of commands without limit
* Per-shell dynamic completion commands

---

## 🚀 Example: Git Clone

```yaml
name: "git"
commands:
  - name: "clone"
    arguments:
      - named: false
        name: "repository"
        completion:
          type: "static"
          values: ["https://github.com/user/repo.git"]
      - named: true
        name: "origin"
        short-name: "o"
```

Generates:

* `git <TAB>` → suggests `clone`.
* `git clone <TAB>` → suggests predefined repository URL.
* `git clone https://... -<TAB>` → suggests `-o` and `--origin`.
