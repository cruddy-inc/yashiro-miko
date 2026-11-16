package domain

type Decision string

const (
	ALLOW   Decision = "ALLOW"
	STEP_UP Decision = "STEP_UP"
	HOLD    Decision = "HOLD"
	BLOCK   Decision = "BLOCK"
)
