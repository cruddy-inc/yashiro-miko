package reroll

type KeyRerollArgs struct{}

func (KeyRerollArgs) Kind() string {
	return "key_reroll"
}
