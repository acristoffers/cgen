package cgen

type Command struct {
	///// Completion /////

	// Command name.
	Name string `yaml:"name"`

	// Other names of this command.
	Aliases []string `yaml:"aliases"`

	// List of accepted arguments.
	Arguments []Argument `yaml:"arguments"`

	// For nesting commands.
	Subcommands []Command `yaml:"commands"`

	///// Manpages /////

	// Deprecated defines, if this command is deprecated. The text will be should in the man page.
	// (should be short)
	Deprecated string `yaml:"deprecated"`

	// If this command is hidden and should NOT show up in the list of available commands.
	Hidden bool `yaml:"hidden"`

	// Usage string. Ex.: add [-F file | -D dir]... [-f format] profile.
	Usage string `yaml:"usage"`

	// Long help message for this command.
	LongDescription string `yaml:"long-description"`

	// Short help message for this command.
	ShortDescription string `yaml:"short-description"`

	// Examples of how to use the command.
	Example string `yaml:"example"`
}
