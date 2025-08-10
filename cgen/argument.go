package cgen

type Argument struct {
	///// Completion /////

	// Whether this is a named or positional argument.
	Named bool `yaml:"named"`

	// If true, options are single-dash: -verbose, and, if false, they are --verbose.
	SingleDashLong bool `yaml:"single-dash-long"`

	// How to separate --long-options from their values: "space", "equal", "both"
	LongValueSeparator string `yaml:"long-value-separator"`

	// How to separate short options (-v) from their values: "space", "attached", "both"
	ShortValueSeparator string `yaml:"short-value-separator"`

	// The name of the argument. If Named==true, for a verbose flag, "verbose" makes "--verbose" or
	// "-verbose". If Named==false, it's the name of the positional argument, e.q.:
	// "configuraton-file" or "container" or "new-name".
	Name string `yaml:"name"`

	// If a short flag version is allowed, its name. (for a verbose flag, "v" makes -v).
	ShortName string `yaml:"short-name"`

	// Short description of the argument validate:"required".
	ShortDescription string `yaml:"short-description"`

	// The completion to suggest.
	Completion Completion `yaml:"completion"`

	///// Manpages /////

	// Deprecated defines, if this command is deprecated. The text will be should in the man page.
	// (should be short)
	Deprecated string `yaml:"deprecated"`

	// Name to display as value placeholder. Will be made UPPERCASE.
	ValueLabel string `yaml:"value-label"`

	// If the arguments should be sorted in the manpage.
	Sort bool `yaml:"sort"`

	// If this argument is hidden and should NOT show up in the list of available arguments.
	Hidden bool `yaml:"hidden"`

	// Long help message for this argument.
	LongDescription string `yaml:"long-description"`

	// Examples of how to use the argument.
	Example string `yaml:"example"`
}

// Provides default arguments
func (i *Argument) UnmarshalYAML(unmarshal func(any) error) error {
	i.Named = false
	i.SingleDashLong = false
	i.LongValueSeparator = "space"
	i.ShortValueSeparator = "space"
	i.Hidden = false
	i.Completion.Type = "none"

	type Arg Argument // prevent recursive call
	return unmarshal((*Arg)(i))
}
