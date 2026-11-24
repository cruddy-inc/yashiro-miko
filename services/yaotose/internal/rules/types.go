package rules

import "cel.dev/cel-go/cel"

type Thresholds struct {
	StepUp int `yaml:"step_up"`
	Hold   int `yaml:"hold"`
	Block  int `yaml:"block"`
}

type Rule struct {
	ID     string `yaml:"id"`
	Force  string `yaml:"force,omitempty"`
	Weight int    `yaml:"weight"`
	When   string `yaml:"when"`
}

type Rules struct {
	Threshold Thresholds `yaml:"thresholds"`
	Rules     []Rule     `yaml:"rules"`
}

type RulesOutcome struct {
	Outcomes Rule
	Program  cel.Program
}

type Engine struct {
	Thresholds Thresholds
	Compiled   []RulesOutcome
}
