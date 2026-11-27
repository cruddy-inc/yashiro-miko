package parameter

import (
	"fmt"
	"strings"
)

type Env string

const (
	Prod    Env = "prod"
	Staging Env = "staging"
)

type Operation string

const (
	Yaotose Operation = "yaotose"
	Tsubaki Operation = "tsubaki"
	Azami   Operation = "azami"
)

type CurrentOperation struct {
	operationName Operation
	environment   Env
}

func NewOperation(name Operation, env Env) (*CurrentOperation, error) {
	switch name {
	case Yaotose, Tsubaki, Azami:
	default:
		return nil, fmt.Errorf("invalid operation %q", name)
	}

	switch env {
	case Prod, Staging:
	default:
		return nil, fmt.Errorf("invalid environment %q", env)
	}

	return &CurrentOperation{
		operationName: name,
		environment:   env,
	}, nil
}

func (c *CurrentOperation) FormatParameter(name string) string {
	cleanName := strings.TrimPrefix(name, "/")
	return fmt.Sprintf("/yashiro-miko/%s/%s/%s", c.operationName, c.environment, cleanName)
}
