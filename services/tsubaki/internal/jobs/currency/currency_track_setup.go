package currency

type CurrencyTrackArgs struct{}

func (CurrencyTrackArgs) Kind() string {
	return "currency_track"
}
