// Package envflag wraps the parsing of the standard flag package to include
// environment variables as a source for flag arguments. Arguments passed as
// command line flags override arguments passed as environment variables.
//
// Clients should call envflag.Parse() instead of flag.Parse().
package envflag

import (
	"flag"
	"os"
	"strings"
)

// An Option is an option.
type Option func(o *option)

type option struct {
	set    *flag.FlagSet
	args   []string
	prefix string
}

// FlagSet returns an Option which specifies the set of flags to parse.
// If unused, flag.CommandLine is the default.
func FlagSet(set *flag.FlagSet) Option {
	return func(o *option) {
		o.set = set
	}
}

// Args returns an Option which specifies the argument list to parse, which
// should not include the command name. If unused, os.Args[1:] is the default.
func Args(arguments []string) Option {
	return func(o *option) {
		o.args = arguments
	}
}

// Prefix returns an Option which specifies a prefix for flag names when
// looking up corresponding enviroment variables.
func Prefix(prefix string) Option {
	return func(o *option) {
		o.prefix = prefix
	}
}

// Parse parses flag definitions from the argument list and the environment,
// giving preference to the argument list over the environment.
func Parse(options ...Option) error {
	o := &option{
		set:  flag.CommandLine,
		args: os.Args[1:],
	}
	for _, opt := range options {
		opt(o)
	}
	if err := o.set.Parse(o.args); err != nil {
		return err
	}
	unset := make(map[string]*flag.Flag)
	o.set.VisitAll(func(f *flag.Flag) { unset[f.Name] = f })
	o.set.Visit(func(f *flag.Flag) { delete(unset, f.Name) })
	var args []string
	for name, f := range unset {
		if v, ok := env(o.prefix + name); ok {
			if isBoolFlag(f.Value) {
				switch strings.ToLower(v) {
				case "true", "yes", "y", "1":
					v = "true"
				case "false", "no", "n", "0":
					v = "false"
				}
			}
			args = append(args, "--"+name+"="+v)
		}
	}
	if len(args) == 0 {
		return nil
	}
	if s := o.set.Args(); len(s) > 0 {
		args = append(append(args, "--"), s...)
	}
	return o.set.Parse(args)
}

func env(name string) (string, bool) {
	key := strings.ToUpper(name)
	key = strings.Replace(key, ".", "_", -1)
	key = strings.Replace(key, "-", "_", -1)
	return os.LookupEnv(key)
}

func isBoolFlag(v flag.Value) bool {
	b, ok := v.(interface{ IsBoolFlag() bool })
	return ok && b.IsBoolFlag()
}
