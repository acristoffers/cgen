package cgen

type Command struct {
	///// Completion /////

	// Command name.
	Name string `yaml:"name"`

	// Other names of this command.
	Aliases []string `yaml:"aliases"`

	// List of accepted arguments.
	Arguments []Argument `yaml:"arguments"`

	// Deprecated defines, if this command is deprecated and should print this string when used.
	Deprecated string `yaml:"deprecated"`

	// For nesting commands.
	Subcommands []Command `yaml:"commands"`

	///// Manpages /////

	// If this command is hidden and should NOT show up in the list of available commands.
	Hidden bool `yaml:"hidden"`

	// Usage string. Ex.: add [-F file | -D dir]... [-f format] profile.
	Usage string `yaml:"usage"`

	// Long help message for this command.
	LongDescription string `yaml:"long-description"`

	// Examples of how to use the command.
	Example string `yaml:"example"`
}
