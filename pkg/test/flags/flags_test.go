// Unit tests for flag parsing.
//
// Eli Bendersky [https://eli.thegreenplace.net]
// This code is in the public domain.
package main

import (
	"flag"
	"reflect"
	"strings"
	"testing"
)

func TestParseFlagsCorrect(t *testing.T) {
	var tests = []struct {
		args []string
		conf Config
	}{
		{[]string{},
			Config{verbose: false, greeting: "", level: 0, args: []string{}}},

		{[]string{"duh"},
			Config{verbose: false, greeting: "", level: 0, args: []string{"duh"}}},

		{[]string{"-verbose"},
			Config{verbose: true, greeting: "", level: 0, args: []string{}}},

		{[]string{"-level", "5"},
			Config{verbose: false, greeting: "", level: 5, args: []string{}}},

		{[]string{"-verbose", "foo", "bar"},
			Config{verbose: true, greeting: "", level: 0, args: []string{"foo", "bar"}}},

		{[]string{"-greeting", "joe"},
			Config{verbose: false, greeting: "joe", level: 0, args: []string{}}},

		{[]string{"-greeting", "joe", "hello"},
			Config{verbose: false, greeting: "joe", level: 0, args: []string{"hello"}}},

		{[]string{"-greeting", "joe", "-verbose"},
			Config{verbose: true, greeting: "joe", level: 0, args: []string{}}},

		{[]string{"-level", "8", "-greeting", "joe", "-verbose", "foo"},
			Config{verbose: true, greeting: "joe", level: 8, args: []string{"foo"}}},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			conf, output, err := parseFlags("prog", tt.args)
			if err != nil {
				t.Errorf("err got %v, want nil", err)
			}
			if output != "" {
				t.Errorf("output got %q, want empty", output)
			}
			if !reflect.DeepEqual(*conf, tt.conf) {
				t.Errorf("conf got %+v, want %+v", *conf, tt.conf)
			}
		})
	}
}

func TestParseFlagsUsage(t *testing.T) {
	var usageArgs = []string{"-help", "-h", "--help"}

	for _, arg := range usageArgs {
		t.Run(arg, func(t *testing.T) {
			conf, output, err := parseFlags("prog", []string{arg})
			if err != flag.ErrHelp {
				t.Errorf("err got %v, want ErrHelp", err)
			}
			if conf != nil {
				t.Errorf("conf got %v, want nil", conf)
			}
			if strings.Index(output, "Usage of") < 0 {
				t.Errorf("output can't find \"Usage of\": %q", output)
			}
		})
	}
}

func TestParseFlagsError(t *testing.T) {
	var tests = []struct {
		args   []string
		errstr string
	}{
		{[]string{"-foo"}, "flag provided but not defined"},
		{[]string{"-level", "joe"}, "invalid value"},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			conf, output, err := parseFlags("prog", tt.args)
			if conf != nil {
				t.Errorf("conf got %v, want nil", conf)
			}
			if strings.Index(err.Error(), tt.errstr) < 0 {
				t.Errorf("err got %q, want to find %q", err.Error(), tt.errstr)
			}
			if strings.Index(output, "Usage of prog") < 0 {
				t.Errorf("output got %q", output)
			}
		})
	}
}
