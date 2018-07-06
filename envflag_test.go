package envflag

import (
	"bytes"
	"flag"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		desc      string
		init      func(*flag.FlagSet)
		args      []string
		env       []string
		prefix    string
		wantFlags map[string]string
		wantArgs  []string
		wantErr   bool
	}{
		{
			desc:      "simple",
			init:      func(f *flag.FlagSet) { f.Int("envflag_simple", 0, "") },
			env:       []string{"ENVFLAG_SIMPLE=42"},
			wantFlags: map[string]string{"envflag_simple": "42"},
		},
		{
			desc:      "prefix",
			init:      func(f *flag.FlagSet) { f.Int("envflag_prefix", 0, "") },
			env:       []string{"PREFIX_ENVFLAG_PREFIX=42"},
			prefix:    "PREFIX_",
			wantFlags: map[string]string{"envflag_prefix": "42"},
		},
		{
			desc:      "args_override_env",
			init:      func(f *flag.FlagSet) { f.Int("envflag_args_override_env", 0, "") },
			args:      []string{"--envflag_args_override_env=42"},
			env:       []string{"ENVFLAG_ARGS_OVERRIDE_ENV=11"},
			wantFlags: map[string]string{"envflag_args_override_env": "42"},
		},
		{
			desc:      "keep_args",
			init:      func(f *flag.FlagSet) { f.Int("envflag_keep_args", 0, "") },
			args:      []string{"keep", "args"},
			env:       []string{"ENVFLAG_KEEP_ARGS=42"},
			wantFlags: map[string]string{"envflag_keep_args": "42"},
			wantArgs:  []string{"keep", "args"},
		},
		{
			desc:      "invalid_arg",
			init:      func(f *flag.FlagSet) { f.Int("envflag_invalid_arg", -1, "") },
			args:      []string{"--envflag_invalid_arg=invalid_int"},
			env:       []string{"ENVFLAG_INVALID_ARG=42"},
			wantFlags: map[string]string{"envflag_invalid_arg": "-1"},
			wantErr:   true,
		},
		{
			desc:      "invalid_env",
			init:      func(f *flag.FlagSet) { f.Int("envflag_invalid_env", -1, "") },
			env:       []string{"ENVFLAG_INVALID_ENV=invalid_int"},
			wantFlags: map[string]string{"envflag_invalid_env": "-1"},
			wantErr:   true,
		},
		{
			desc:      "invalid_bool",
			init:      func(f *flag.FlagSet) { f.Bool("bool", false, "") },
			env:       []string{"BOOL=nope"},
			wantFlags: map[string]string{"bool": "nope"},
			wantErr:   true,
		},
		{
			desc: "lowercase_bool",
			init: func(f *flag.FlagSet) {
				f.Bool("true", false, "")
				f.Bool("yes", false, "")
				f.Bool("y", false, "")
				f.Bool("1", false, "")
				f.Bool("false", true, "")
				f.Bool("no", true, "")
				f.Bool("n", true, "")
				f.Bool("0", true, "")
			},
			env: []string{
				"TRUE=true",
				"YES=yes",
				"Y=y",
				"1=1",
				"FALSE=false",
				"NO=no",
				"N=n",
				"0=0",
			},
			wantFlags: map[string]string{
				"true":  "true",
				"yes":   "true",
				"y":     "true",
				"1":     "true",
				"false": "false",
				"no":    "false",
				"n":     "false",
				"0":     "false",
			},
		},
		{
			desc: "uppercase_bool",
			init: func(f *flag.FlagSet) {
				f.Bool("true", false, "")
				f.Bool("yes", false, "")
				f.Bool("y", false, "")
				f.Bool("1", false, "")
				f.Bool("false", true, "")
				f.Bool("no", true, "")
				f.Bool("n", true, "")
				f.Bool("0", true, "")
			},
			env: []string{
				"TRUE=TRUE",
				"YES=YES",
				"Y=Y",
				"1=1",
				"FALSE=FALSE",
				"NO=NO",
				"N=N",
				"0=0",
			},
			wantFlags: map[string]string{
				"true":  "true",
				"yes":   "true",
				"y":     "true",
				"1":     "true",
				"false": "false",
				"no":    "false",
				"n":     "false",
				"0":     "false",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			defer resetEnv()()
			setEnv(tt.env)
			set := flag.NewFlagSet(tt.desc, flag.ContinueOnError)
			tt.init(set)
			w := bytes.NewBuffer(nil)
			set.SetOutput(w)
			opts := []Option{FlagSet(set), Args(tt.args)}
			if tt.prefix != "" {
				opts = append(opts, Prefix(tt.prefix))
			}
			if err := Parse(opts...); err != nil {
				if !tt.wantErr {
					t.Logf("Output:\n%s", w.Bytes())
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("expected error")
			}
			flags := make(map[string]string)
			set.VisitAll(func(f *flag.Flag) { flags[f.Name] = f.Value.String() })
			if !reflect.DeepEqual(flags, tt.wantFlags) {
				t.Errorf("flags: want: %v; got: %v", tt.wantFlags, flags)
			}
			if args := set.Args(); len(args)+len(tt.wantArgs) > 0 && !reflect.DeepEqual(tt.wantArgs, args) {
				t.Errorf("args: want: %v; got: %v", tt.wantArgs, args)
			}
		})
	}
}

func resetEnv() func() {
	env := os.Environ()
	os.Clearenv()
	return func() {
		os.Clearenv()
		setEnv(env)
	}
}

func setEnv(env []string) {
	for _, s := range env {
		kv := strings.SplitN(s, "=", 2)
		os.Setenv(kv[0], kv[1])
	}
}
