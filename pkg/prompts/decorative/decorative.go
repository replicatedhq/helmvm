// Package decorative implement decorative prompts using the survey library.
package decorative

import (
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/sirupsen/logrus"
)

// Decorative is a decorative prompt.
type Decorative struct {
	in  terminal.FileReader
	out terminal.FileWriter
}

type Option func(p *Decorative)

func New(opts ...Option) Decorative {
	p := Decorative{}
	for _, opt := range opts {
		opt(&p)
	}
	if p.in == nil {
		p.in = os.Stdin
	}
	if p.out == nil {
		p.out = os.Stdout
	}
	return p
}

func WithIn(in terminal.FileReader) Option {
	return func(p *Decorative) {
		p.in = in
	}
}

func WithOut(out terminal.FileWriter) Option {
	return func(p *Decorative) {
		p.out = out
	}
}

// Confirm asks for user for a "Yes" or "No" response. The default value
// is used if the user presses enter without typing anything.
func (d Decorative) Confirm(msg string, defvalue bool) bool {
	var response bool
	var confirm = &survey.Confirm{Message: msg, Default: defvalue}
	confirm.WithStdio(d.stdio())
	if err := survey.AskOne(confirm, &response); err != nil {
		logrus.Fatalf("unable to confirm: %v", err)
	}
	return response
}

// PressEnter asks the user to press enter to continue.
func (d Decorative) PressEnter(msg string) {
	var i string
	in := &survey.Input{Message: msg}
	in.WithStdio(d.stdio())
	if err := survey.AskOne(in, &i); err != nil {
		logrus.Fatalf("unable to ask for input: %v", err)
	}
}

// Password asks the user for a password. Password can't be empty.
func (d Decorative) Password(msg string) string {
	var pass string
	for pass == "" {
		question := &survey.Password{Message: msg}
		question.WithStdio(d.stdio())
		if err := survey.AskOne(question, &pass); err != nil {
			logrus.Fatalf("unable to ask for input: %v", err)
		} else if pass == "" {
			logrus.Error("Password cannot be empty")
		}
	}
	return pass
}

// Select asks the user to select one of the provided options.
func (d Decorative) Select(msg string, options []string, defvalue string) string {
	question := &survey.Select{
		Message: msg,
		Options: options,
		Default: defvalue,
	}
	question.WithStdio(d.stdio())
	var response string
	if err := survey.AskOne(question, &response); err != nil {
		logrus.Fatalf("unable to ask for input: %v", err)
	}
	return response
}

// Input asks the user for a string. If required is true then
// the string cannot be empty.
func (d Decorative) Input(msg string, defvalue string, required bool) string {
	var response string
	for response == "" {
		question := &survey.Input{Message: msg, Default: defvalue}
		question.WithStdio(d.stdio())
		if err := survey.AskOne(question, &response); err != nil {
			logrus.Fatalf("unable to ask for input: %v", err)
		} else if !required || response != "" {
			break
		}
		logrus.Error("Input cannot be empty")
	}
	return response
}

func (d Decorative) stdio() terminal.Stdio {
	return terminal.Stdio{
		In:  d.in,
		Out: d.out,
	}
}
