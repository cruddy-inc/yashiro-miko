package kafka

type CasesOpenedOutboxArgs struct{}

func (CasesOpenedOutboxArgs) Kind() string {
	return "cases_opened_outbox"
}
