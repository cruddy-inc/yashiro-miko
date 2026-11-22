---
id: kb-bank-impersonation-safe-account-001
title: Safe Account Transfer Scam
typology: bank_impersonation
sub_typology: safe_account_transfer
status: draft
version: "0.1"
last_updated: "2026-11-22"
rails:
  - SEPA_Instant
signals:
  - active_phone_call
  - remote_access_detected
  - known_payee
  - payee_verification
  - amount_minor
  - payer_max_amount_minor_90d
  - payer_out_count_1h
  - payee_in_count_24h
tools:
  - GetPayeeContext
  - GetNeighborhood
  - GetPayerSummary
severity_range: STEP_UP-HOLD
interventions:
  - WARN_SAFE_ACCOUNT
  - COOLING_OFF
  - CALL_BACK
sources:
  - title: Starling, safe account scams
    url: https://starlingbank.com/docs/security/safe-account-scams.pdf
    accessed: "2026-11-22"
  - title: Cifas, remote access scam
    url: https://www.cifas.org.uk/newsroom/fraudsters-steal-millions-from-uk-businesses-through-remote-access-bank-scam
    accessed: "2026-11-22"
  - title: FCA, mule cash-out
    url: https://www.fca.org.uk/publications/multi-firm-reviews/money-mules-activity-cashing-out-findings
    accessed: "2026-11-22"
---

# Safe Account Transfer Scam

## How it works

A safe account transfer scam starts with a caller posing as the victim's bank fraud team or the police, often after phishing. The caller claims the account is compromised and the money is safe only if moved to a "safe" account, which the fraudster controls. The victim stays on the line, is coached through each screen and told to hide the reason from bank staff. Urgency, authority and secrecy keep them compliant; some also install remote access software or read out one-time codes. The victim authorizes the payment, so credentials and strong authentication pass.

## What it looks like in the data

| Signal | Usual value | Suggests | Strength |
|---|---|---|---|
| `active_phone_call` | true | live coaching | supporting |
| `remote_access_detected` | true | screen control | strong |
| `known_payee` | false | new payee | supporting |
| `payee_verification` | no match or unavailable | holder not as believed | supporting |
| `amount_minor` vs `payer_max_amount_minor_90d` | far above usual maximum | out of pattern | strong in combination |
| `payer_out_count_1h` | quick payments or retries | payer being guided | supporting |
| `payee_in_count_24h` | many unrelated payers | collection account | strong |

Combinations outweigh single rows. A new payee, an out-of-pattern amount and a live call together suggest coaching; remote access on top is strongest. The data holds no call content, and a payer coached on a landline or second phone can show no call signal.

## Investigation

1. Read the case features first; the table rows need no tool call.
2. `GetPayeeContext`: account age, distinct payers in 24h, outflow ratio, confirmed mule flag.
3. `GetNeighborhood`, only if inflow looks high: fan-in from unrelated payers, speed of onward movement.
4. `GetPayerSummary`: new device, recent limit increase, new payee.

**Worrying results:** payee younger than <N> days with <N>+ distinct payers in 24h; most inbound money leaving within <N> minutes.

**Reassuring results:** established payee, verified name match, low fan-in.

## Look-alikes that are usually fine

- **Genuine bank call:** no transfer to a new account is requested.
- **Own account elsewhere:** `payee_verification` matches the payer's name.
- **Verified seller or solicitor:** established payee, name match.
- **IT support call:** remote access present, but payee and amount are routine for the payer.

## Why it's risky

Instant payments settle in seconds and are hard to recall. Funds are split and moved onward, often within minutes.

## Typical severity and intervention

| Pattern | Severity | Intervention |
|---|---|---|
| one supporting signal alone | STEP_UP | WARN_SAFE_ACCOUNT |
| active call or remote access, plus new payee and out-of-pattern amount | HOLD | COOLING_OFF or CALL_BACK |
| new payee, high fan-in or fast outflow, no call or remote access | STEP_UP to HOLD | COOLING_OFF |

## Example

```
yaotose_decision: "STEP_UP"
current_transaction_info {
  segment: "retail"
  original_currency: "EUR"
  amount_eur_minor: 480000
  active_phone_call: true
  remote_access_detected: true
}
payer_to_payee_interaction {
  known_payee: false
  n_payments: 0
}
```

Reading: EUR 4,800 to a payee with no history, live call plus remote access: coached payment.