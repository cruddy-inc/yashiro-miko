## Main Purpose

This file explains every table in our "mock bank" ([oumi](https://github.com/cruddy-inc/yashiro-miko/tree/main/services/oumi)) and the reasoning behind it. All tables were generated with AI, because there is no public, universal reference for where a real bank stores its customers, payments, history and so on. The only option we had was to mock them.

## Global patterns

- **Schema** -> Everything lives in the `bank` schema (`CREATE SCHEMA IF NOT EXISTS bank;`). This keeps PostgreSQL from putting everything in `public`. That isn't necessarily bad, but in bigger projects with multiple schemas it gets messy fast.
- **Money as minor units** -> All money fields use the smallest currency unit (cents), to avoid floating-point rounding errors.
- **Sensitive data** -> The mock bank stores fake data in plain text. `yaotose` never works with raw identifiers: IBANs are tokenized (keyed HMAC) before analysis, and raw values are kept only encrypted in its vault.
- **Constant REGEX checks** -> Several fields use REGEX checks to match real-world formats, such as the country code (`ISO 3166-1 alpha-2`) or the currency (`ISO 4217`).
- **Risk factors** -> We don't check fields such as `kyc_status` or `is_pep`. Those are bank gates, and `yaotose` checks the context of a payment, not bank gates.

## 1. Customer Profile (Table)

The first table holds customer profiles and KYC (Know Your Customer) records. It stores what we need to know about a customer under **Know Your Customer (KYC)** and **Anti-Money Laundering (AML)** rules. Some of these fields provide context that can help flag a payment as `STEP_UP`, `HOLD` or `BLOCK`, for example a brand-new customer sending a large amount. Design decisions made with AI assistance:

- **What it stores** -> Data about both individuals and companies (SME).
- **Why this choice** -> Instead of two separate tables with many redundant columns, `date_of_birth` is nullable, since companies have no birth date.
- **Risk signals:**
    - **`kyc_risk_rating`** -> Rating is "HIGH".
    - **`onboarded_at`** -> New customer + large amount: scam/mule signal.
    - **`segment`** -> Decides what a normal amount is.
    - **`country`** -> Login from a foreign country: account-takeover signal.
- REGEX used, and why:
    - **`country`:**
        - **REGEX** -> `~ '^[A-Z]{2}$'`
        - **MEANING** -> Checks so the format contains two uppercase letters, as in ISO 3166-1 alpha-2 country codes (IT, RO, AU), the same codes that start an IBAN ([list on iban.com](https://www.iban.com/country-codes)).
    - **`email`:**
        - **REGEX** -> `~* '^[a-z0-9._+-]{1,64}@([a-z0-9-]{1,63}\.)+([a-z0-9]{2,18})$'`
        - **MEANING** -> `~*` makes the match case-insensitive. The address must follow the usual format (`example@gmail.com`, `yashiro_miko.idk+idk@outlook.com`). This is a simplified check, not a full email validator.
    - **`phone_e164`:**
        - **REGEX** -> `~ '^\+[1-9][0-9]{6,14}$'`
        - **MEANING** -> Must start with `+`, followed by a non-zero digit and then 6 to 14 more digits (`+40700000001`, `+14155550100`).

## 2. Bank Account (Table)

The second table stores a customer's bank accounts: IBAN, currency, balances, limits and account type. It is a mock of what a real bank account could hold. Some checks are deliberately left out, such as limits for minors, because the bank already enforces them elsewhere, so `yaotose` can safely ignore them.

- **What it stores** -> The IBAN, currency, balances, limits and type of each account. It holds no card data.
- **Why this choice** -> Instead of piling everything into a single table (customer + banking data), accounts live in their own table with a parent/child (one-to-many) relationship: one customer can own several accounts.
- **Risk signals:**
    - **`currency` + `available_balance_minor`** -> A payment that takes most of the available balance is a fraud signal.
    - **`account_type`** -> Activity that doesn't fit the type: a "CURRENT" account receiving payments as if it were "BUSINESS", or frequent transfers between "SAVINGS" and "CURRENT".
    - **`opened_at`** -> High-amount or high-frequency transactions on a recently opened account.
- REGEX used, and why:
    - **`iban`:**
        - **REGEX** -> `~ '^[A-Z]{2}[0-9]{2}[A-Z0-9]{11,30}$'`
        - **MEANING** -> Two uppercase letters (country), two digits (check digits), then 11 to 30 letters or digits, which gives the 15 to 34 characters allowed by ISO 13616. This checks the format only. The mod-97 checksum and the country-specific length are validated in yaotose.
    - **`currency`:**
        - **REGEX** -> `~ '^[A-Z]{3}$'`
        - **MEANING** -> Three uppercase letters, as in ISO 4217 currency codes (EUR, RON, USD). This checks the format only, not that the code exists.

## 3. Customer Device (Table)

This table stores information about the devices a customer uses to log in to their account. It helps reveal major breaches such as account takeover (ATO) and several other attacker strategies.

- **What it stores** -> The devices connected to an account and several details about each one.
- **Why this choice** -> Knowing which devices are connected to a customer matters. If a new device appears on an account and starts sending many transactions, we can act on it.
- **Risk signals:**
    - **`first_seen_at`** -> A new device plus a burst of transactions = ATO risk.
    - **`bound_at`** -> Attackers usually bind their own device to the account and then drain it.

## 4. Login Sessions (Table)

This table stores the customer's login history: each session, the device and network it came from, and how the customer authenticated. It helps us spot a new device suddenly logging in to an account and see what it did afterwards.

- **What it stores** -> Each login session of a customer, with details about the device, the network (IP, country, proxy) and the authentication method.
- **Why this choice** -> It matters how a customer signs in (PIN, biometrics, password + 2FA), so we can tell a normal session from a suspicious one.
- **Risk signals:**
    - **`ip_country` / `is_proxy`** -> A different country + a different device + a proxy = HIGH RISK.
    - **`auth_method`** -> A change from the customer's usual method (for example, biometrics one day, password the next) is a takeover signal.
- REGEX used, and why:
    - **`ip_country`:**
        - **REGEX** -> `~ '^[A-Z]{2}$'`
        - **MEANING** -> Checks the format only: two uppercase letters, as in ISO 3166-1 alpha-2 country codes (IT, RO, AU). It does not check that the code actually exists.

## 5. Authentication Events (Table)

This table is an append-only log of security events on an account: successful and failed logins, failed MFA checks, password resets and device bindings. A session (section 4) exists only after a successful login, while this log also records the failures, which is where most attacks leave their first trace.

- **What it stores** -> One row per authentication event: its type, time, device and IP/country.
- **Why this choice** -> It is separate from `sessions` because a session is a successful login with a lifetime, while an event is a single moment. Failed attempts never create a session, but they are exactly what we want to count.
- **Risk signals:**
    - **`LOGIN_FAILED` / `MFA_FAILED`** -> Many failures in a short window, followed by a success = password guessing or phishing that worked. Matsuri turns this into `failed_logins_24h`.
    - **`PASSWORD_RESET`** -> A reset shortly before a large payment is an account-takeover signal.
    - **`DEVICE_BOUND`** -> A new device bound minutes before a payment (also stored in `devices.bound_at`).
- The REGEX on `ip_country` is the same format check explained in section 1.

## 6. Profile Changes (Table)

This table logs changes to a customer's phone, email, address and limits. Attackers often change the contact details first, so the victim never sees the alerts or OTP codes, or raise the limits before draining the account.

- **What it stores** -> Which field changed, when, from which session, and the old and new values as hashes.
- **Why this choice** -> A dedicated change log keeps a clear timeline next to the `customers` table. Only hashes are stored, so the log itself can't leak an old phone number or email.
- **Risk signals:**
    - **`PHONE` / `EMAIL`** -> Contact details changed shortly before a large payment = account-takeover signal.
    - **`LIMIT`** -> A limit raised shortly before a large payment.
    - **Combined** -> A profile change + a new device + a new payee within the same hour is a much stronger signal than any one alone.
- The only constraint on the hash columns is a minimum length (32 characters), not a REGEX.

## 7. Beneficiaries (Table)

This table holds the payees a customer has saved, together with the Verification of Payee (VoP) result: the bank's check of whether the name entered matches the name on the account.

- **What it stores** -> The saved payee (name, IBAN, BIC, nickname), when and from which session it was added, the VoP result and when it was last used.
- **Why this choice** -> A payee the customer has used for months is very different from one added a minute ago. Keeping this as its own table lets us tell a saved payee from a one-off payment, and `UNIQUE (customer_id, iban)` prevents duplicate entries.
- **Risk signals:**
    - **`added_at`** -> A recently added payee + a large amount = scam signal.
    - **`vop_result`** -> `NO_MATCH` or `CLOSE_MATCH` means the name doesn't match the account holder. The bank runs the check, and `yaotose` only uses the result as a signal.
    - **No row at all** -> A payee that was never saved at all.
- REGEX used, and why:
    - **`iban`:** same format check as section 2.
    - **`bic`:**
        - **REGEX** -> `~ '^[A-Z]{4}[A-Z]{2}[A-Z0-9]{2}([A-Z0-9]{3})?$'`
        - **MEANING** -> The ISO 9362 BIC format: a 4-letter bank code, a 2-letter country code, a 2-character location code and an optional 3-character branch code (8 or 11 characters in total). It checks the format only. The same pattern is reused in sections 8 and 9.

## 8. Payment Orders (Table)

This table stores outgoing payment instructions, from the moment the customer submits them until they are executed, held or rejected. It is where a payment request starts, and the request sent to `yaotose` is built from one row of it.

- **What it stores** -> The creditor (name, IBAN, BIC), the amount and currency, the payment scheme, the SEPA references, the session that initiated it, and its status with a reason.
- **Why this choice** -> It is separate from `transactions` on purpose: an order is an instruction with a life cycle (RECEIVED, SCREENING, HELD, EXECUTED, REJECTED, RETURNED), while a transaction is a ledger entry that exists only after the money has moved. After screening, matsuri writes the decision back into `status`.
- **Risk signals:**
    - **`amount_minor`** -> Far above the customer's usual amount.
    - **`scheme`** -> `SEPA_INSTANT` can't be recalled, so the stakes are higher.
    - **`remittance_info`** -> Scam wording, or injected instructions aimed at the AI analyst (this free text is handled by the quarantined path).
    - **`creditor_iban` / `creditor_name`** -> A payee with many senders, or a name that doesn't match.
- Length limits from the SEPA rules (checks, not REGEX): `end_to_end_id` up to 35 characters and `remittance_info` up to 140 characters. IBAN, BIC and currency use the formats explained earlier.

## 9. Transactions (Table)

This table is the settled account statement: the ledger entries that exist once money has actually moved, as the customer sees them. Real cores use double-entry postings, so this is a simplified, per-account view.

- **What it stores** -> Booked debits and credits with the amount, balance after the entry, scheme, counterparty and references.
- **Why this choice** -> It is the bank's history. Keeping it apart from `payment_orders` gives a clear split between "what the customer asked for" and "what actually happened".
- **Risk signals** (these are computed by `yaotose` from its own copy, not read from this table):
    - **`counterparty_iban`** -> Has this customer paid this payee before?
    - **`amount_minor`** -> The customer's usual amount (median and 95th percentile over recent months).
    - **`booked_at`** -> How many payments in the last 10 minutes or 24 hours.
- `yaotose` never queries this table. `matsuri sync-history` exports the debits to `yaotose`, which keeps its own tokenized history. Ledger correctness (for example `balance_after_minor`) is the bank's bookkeeping, and `yaotose` doesn't check it.
- No new REGEX: it reuses the IBAN, BIC and currency formats from the earlier sections.