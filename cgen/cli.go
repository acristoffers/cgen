package cgen

type CLI struct {
	// The tool's name.
	Name string `yaml:"name" validate:"required"`

	// Tool's short description.
	ShortDescription string `yaml:"short-description"`

	// Tool's long description.
	LongDescription string `yaml:"long-description"`

	// Tool's version.
	Version string `yaml:"version"`

	// Global arguments.
	Arguments []Argument `yaml:"arguments"`

	// Top-level Commands
	Commands []Command `yaml:"commands"`
}
