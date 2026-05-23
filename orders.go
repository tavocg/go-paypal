package paypal

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	createOrderEndpoint         = "/v2/checkout/orders"
	captureOrderPaymentEndpoint = "/v2/checkout/orders/{orderID}/capture"
)

type CreateOrderRequest struct {
	PaymentSource struct {
		Paypal struct {
			ExperienceContext struct {
				PaymentMethodPreference string `json:"payment_method_preference"`
				LandingPage             string `json:"landing_page"`
				ShippingPreference      string `json:"shipping_preference"`
				UserAction              string `json:"user_action"`
				ReturnURL               string `json:"return_url"`
				CancelURL               string `json:"cancel_url"`
			} `json:"experience_context"`
		} `json:"paypal"`
	} `json:"payment_source"`
	PurchaseUnits []struct {
		InvoiceID string `json:"invoice_id"`
		Amount    struct {
			CurrencyCode string `json:"currency_code"`
			Value        string `json:"value"`
			Breakdown    struct {
				ItemTotal struct {
					CurrencyCode string `json:"currency_code"`
					Value        string `json:"value"`
				} `json:"item_total"`
				Shipping struct {
					CurrencyCode string `json:"currency_code"`
					Value        string `json:"value"`
				} `json:"shipping"`
			} `json:"breakdown"`
		} `json:"amount"`
		Items []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			UnitAmount  struct {
				CurrencyCode string `json:"currency_code"`
				Value        string `json:"value"`
			} `json:"unit_amount"`
			Quantity string `json:"quantity"`
			Category string `json:"category"`
			Sku      string `json:"sku"`
			ImageURL string `json:"image_url"`
			URL      string `json:"url"`
			Upc      struct {
				Type string `json:"type"`
				Code string `json:"code"`
			} `json:"upc"`
		} `json:"items"`
	} `json:"purchase_units"`
}

type CreateOrderResponse struct {
	ID            string `json:"id"`
	PaymentSource struct {
		Paypal struct{} `json:"paypal"`
	} `json:"payment_source"`
	Links []struct {
		Href   string `json:"href"`
		Rel    string `json:"rel"`
		Method string `json:"method"`
	} `json:"links"`
}

type CaptureOrderPaymentResponse struct {
	ID            string `json:"id"`
	PaymentSource struct {
		Paypal struct {
			Name struct {
				GivenName string `json:"given_name"`
				Surname   string `json:"surname"`
			} `json:"name"`
			EmailAddress string `json:"email_address"`
			AccountID    string `json:"account_id"`
		} `json:"paypal"`
	} `json:"payment_source"`
	PurchaseUnits []struct {
		ReferenceID string `json:"reference_id"`
		Shipping    struct {
			Address struct {
				AddressLine1 string `json:"address_line_1"`
				AddressLine2 string `json:"address_line_2"`
				AdminArea2   string `json:"admin_area_2"`
				AdminArea1   string `json:"admin_area_1"`
				PostalCode   string `json:"postal_code"`
				CountryCode  string `json:"country_code"`
			} `json:"address"`
		} `json:"shipping"`
		Payments struct {
			Captures []struct {
				ID     string `json:"id"`
				Status string `json:"status"`
				Amount struct {
					CurrencyCode string `json:"currency_code"`
					Value        string `json:"value"`
				} `json:"amount"`
				SellerProtection struct {
					Status            string   `json:"status"`
					DisputeCategories []string `json:"dispute_categories"`
				} `json:"seller_protection"`
				FinalCapture              bool `json:"final_capture"`
				SellerReceivableBreakdown struct {
					GrossAmount struct {
						CurrencyCode string `json:"currency_code"`
						Value        string `json:"value"`
					} `json:"gross_amount"`
					PaypalFee struct {
						CurrencyCode string `json:"currency_code"`
						Value        string `json:"value"`
					} `json:"paypal_fee"`
					NetAmount struct {
						CurrencyCode string `json:"currency_code"`
						Value        string `json:"value"`
					} `json:"net_amount"`
				} `json:"seller_receivable_breakdown"`
				CreateTime time.Time `json:"create_time"`
				UpdateTime time.Time `json:"update_time"`
				Links      []struct {
					Href   string `json:"href"`
					Rel    string `json:"rel"`
					Method string `json:"method"`
				} `json:"links"`
			} `json:"captures"`
		} `json:"payments"`
	} `json:"purchase_units"`
	Payer struct {
		Name struct {
			GivenName string `json:"given_name"`
			Surname   string `json:"surname"`
		} `json:"name"`
		EmailAddress string `json:"email_address"`
		PayerID      string `json:"payer_id"`
	} `json:"payer"`
	Links []struct {
		Href   string `json:"href"`
		Rel    string `json:"rel"`
		Method string `json:"method"`
	} `json:"links"`
}

func (c *Client) CreateOrder(ctx context.Context, order CreateOrderRequest) (*CreateOrderResponse, error) {
	response := &CreateOrderResponse{}
	if err := c.request(
		ctx,
		http.MethodPost,
		createOrderEndpoint,
		order,
		response,
	); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *Client) CaptureOrderPayment(ctx context.Context, orderID string) (*CaptureOrderPaymentResponse, error) {
	if orderID == "" {
		return nil, fmt.Errorf("orderID is required")
	}

	response := &CaptureOrderPaymentResponse{}
	endpoint := strings.Replace(captureOrderPaymentEndpoint, "{orderID}", orderID, 1)

	if err := c.request(
		ctx,
		http.MethodPost,
		endpoint,
		nil,
		response,
	); err != nil {
		return nil, err
	}
	return response, nil
}
