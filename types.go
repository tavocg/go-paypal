package paypal

// Shared types that exactly match the responses of multiple flows in PayPal.
type Money struct {
	CurrencyCode string `json:"currency_code"`
	Value        string `json:"value"`
}

type PostalAddress struct {
	AddressLine1 string `json:"address_line_1,omitempty"`
	AddressLine2 string `json:"address_line_2,omitempty"`
	AdminArea2   string `json:"admin_area_2,omitempty"`
	AdminArea1   string `json:"admin_area_1,omitempty"`
	PostalCode   string `json:"postal_code,omitempty"`
	CountryCode  string `json:"country_code"`
}
