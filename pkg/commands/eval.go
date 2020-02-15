package commands

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

const valuesHashName = "values"

type EvalCommand struct {
	Writer    io.Writer
	Template  string   `short:"t" long:"template" description:"path to yaml template you would like to render"`
	Values    []string `short:"c" long:"values" description:"path to values file(s) you would like to use for rendering"`
	Policy    string   `short:"p" long:"policy" description:"path to rego policies to evaluate against rendered templates"`
	Namespace string   `short:"n" long:"namespace" description:"policy namespace to query for rules"`
	Verbose   bool     `short:"v" long:"verbose" description:"prints tracing output to stdout"`
}

func (s *EvalCommand) Execute(args []string) error {
	s.setDefaults()

	if s.Policy == "" {
		return InvalidPolicyPath
	}

	fileFile, err := os.Open(s.Policy)
	if err != nil {
		return InvalidPolicyPath
	}
	fileFile.Close()
	valuesConfig, err := mergeValues(s.Values)
	if err != nil {
		return fmt.Errorf("failed merging values files %w ", err)
	}

	renderedOutput, err := validateAndRender(s.Template, valuesConfig)
	if err != nil {
		return fmt.Errorf("error while rendering: %w", err)
	}

	policyInput, err := UnmarshalYamlMap(renderedOutput)
	if err != nil {
		return fmt.Errorf("formatting policy input failed: %w", err)
	}

	policyInput[valuesHashName] = valuesConfig
	return evalPolicyOnInput(s.Writer, s.Policy, s.Namespace, policyInput)
}

func (s *EvalCommand) setDefaults() {
	if s.Writer == nil {
		s.Writer = os.Stdout
	}

	if !s.Verbose {
		s.Writer = new(bytes.Buffer)
	}

	if s.Namespace == "" {
		s.Namespace = "main"
	}
}
