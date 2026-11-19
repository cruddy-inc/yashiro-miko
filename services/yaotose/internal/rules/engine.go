package rules

import (
	"fmt"
	"os"

	"cel.dev/cel-go/cel"
	"github.com/goccy/go-yaml"
)

func NewEngine() (*Engine, error) {
	e, err := cel.NewEnv(
		cel.Variable("remote_access_detected", cel.BoolType),
		cel.Variable("active_phone_call", cel.BoolType),
		cel.Variable("known_payee", cel.BoolType),
		cel.Variable("amount_eur_minor", cel.IntType),
		cel.Variable("payee_verification", cel.StringType),
		cel.Variable("n_payments", cel.IntType),
		cel.Variable("max_minor", cel.IntType),
		cel.Variable("now", cel.IntType),
		cel.Variable("first_transaction", cel.IntType),
		cel.Variable("last_transaction", cel.IntType),
		cel.Variable("segment", cel.StringType),
		cel.Variable("count_1h", cel.IntType),
		cel.Variable("sum_1h", cel.IntType),
		cel.Variable("count_24h", cel.IntType),
		cel.Variable("sum_24h", cel.IntType),
	)
	if err != nil {
		return nil, err
	}

	fc, err := os.ReadFile("rules.yaml")
	if err != nil {
		return nil, err
	}

	var con Rules
	if err = yaml.Unmarshal(fc, &con); err != nil {
		return nil, err
	}

	compiled := make([]RulesOutcome, 0, len(con.Rules))
	for i := 0; i < len(con.Rules); i++ {
		b := con.Rules[i]

		ast, issue := e.Compile(b.When)
		if issue != nil && issue.Err() != nil {
			return nil, issue.Err()
		}

		if ast.OutputType() != cel.BoolType {
			return nil, fmt.Errorf("rule %q must return boolean", b.When)
		}

		prg, err := e.Program(ast)
		if err != nil {
			return nil, err
		}

		compiled = append(compiled, RulesOutcome{
			Outcomes: b,
			Program:  prg,
		})
	}

	return &Engine{
		Thresholds: con.Threshold,
		Compiled:   compiled,
	}, nil
}
