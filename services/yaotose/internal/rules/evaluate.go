package rules

import "yaotose/internal/domain"

func (e *Engine) Evaluate(assingVariables map[string]any) (*domain.Verdict, error) {
	var matchRule []string
	var weight int
	for i := 0; i < len(e.Compiled); i++ {
		b := e.Compiled[i]

		out, _, err := b.Program.Eval(assingVariables)
		if err != nil {
			return nil, err
		}

		// Skip not matching 'when's
		if out.Value() != true {
			continue
		}

		matchRule = append(matchRule, b.Outcomes.ID)
		weight += b.Outcomes.Weight

		if b.Outcomes.Force == "HOLD" {
			return &domain.Verdict{
				SuspicionAlert: []string{b.Outcomes.ID},
				Weight:         b.Outcomes.Weight,
				Decision:       "HOLD",
			}, nil
		}
	}

	decision := "ALLOW"
	if weight >= e.Thresholds.StepUp {
		decision = "STEP_UP"
	}
	if weight >= e.Thresholds.Hold {
		decision = "HOLD"
	}
	if weight >= e.Thresholds.Block {
		decision = "BLOCK"
	}

	return &domain.Verdict{
		SuspicionAlert: matchRule,
		Decision:       decision,
		Weight:         weight,
	}, nil
}
