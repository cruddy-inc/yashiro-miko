package domain

type IncomingPayment struct {
	IntentID             string `json:"intent_id"`
	TenantID             string `json:"tenant_id"`
	Amount               string `json:"amount"`
	Currency             string `json:"currency"`
	PayerIban            string `json:"payer_iban"`
	PayeeIban            string `json:"payee_iban"`
	PayeeName            string `json:"payee_name"`
	Segment              string `json:"segment"`
	RemittanceInfo       string `json:"remittance_info"`
	PayeeVerification    string `json:"payee_verification"`
	ActivePhoneCall      bool   `json:"active_phone_call"`
	RemoteAccessDetected bool   `json:"remote_access_detected"`
}

type SanitizedPayment struct {
	IntentID             string
	TenantID             string
	AmountCents          int64
	Currency             string
	PayerToken           string
	PayeeToken           string
	RemittanceInfo       string
	PayeeVerification    string
	ActivePhoneCall      bool
	RemoteAccessDetected bool
}

type RecipientOutcome struct {
	IntentID string
	Decision Decision
	Type     []string
}
