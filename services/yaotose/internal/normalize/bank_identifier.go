package normalize

func GetBankIdentifier(ibanNormalized string) string {
	switch ibanNormalized[0:2] {
	case "EE", "FK", "GE", "IS", "MD", "SA", "SD", "TN", "XK":
		return ibanNormalized[4:6]
	case "AE", "AL", "BA", "BE", "CY", "FI", "GR", "HU", "IL", "KZ", "LU", "LY", "ME", "MK", "OM", "RS", "SE", "TL", "VA":
		return ibanNormalized[4:7]
	case "AD", "AZ", "BG", "BH", "BY", "CR", "CZ", "DK", "DO", "EG", "ES", "FO", "GB", "GI", "GL", "GT", "HN", "IE", "IQ", "JO", "KW", "LB", "LC", "LV", "MN", "MT", "NI", "NL", "NO", "PK", "PS", "PT", "QA", "RO", "SK", "SO", "ST", "SV", "VG", "YE":
		return ibanNormalized[4:8]
	case "AT", "BI", "CH", "DJ", "FR", "LI", "LT", "MC", "MR", "SI", "TR":
		return ibanNormalized[4:9]
	case "MU", "SC", "UA":
		return ibanNormalized[4:10]
	case "HR":
		return ibanNormalized[4:11]
	case "BR", "DE", "PL":
		return ibanNormalized[4:12]
	case "RU":
		return ibanNormalized[4:13]
	case "IT", "SM":
		return ibanNormalized[5:10]
	default:
		return ""
	}
}
