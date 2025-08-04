package cgen

type Completion struct {
	// One of "function", "static", "none", "file", "folder"
	// Function: uses the return of Fish, Bash and Zsh as completion
	// Static: uses the values in Values
	// File: complete with a file name
	// Folder: complete with a folder name
	// None: the argument takes a value, but cannot be autocompleted, so suggest nothing
	Type string `yaml:"type" validate:"oneof=function static none file folder"`

	// Fish code to return completions.
	Fish string `yaml:"fish"`

	// Bash code to return completions.
	Bash string `yaml:"bash"`

	// Zsh code to return completions.
	Zsh string `yaml:"zsh"`

	// Static list of values to suggest.
	Values []string `yaml:"values"`
}

// Provides default arguments
func (i *Completion) UnmarshalYAML(unmarshal func(any) error) error {
	i.Type = "none"

	type Comp Completion // prevent recursive call
	return unmarshal((*Comp)(i))
}
