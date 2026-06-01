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

type Order struct {
	Intent        string                     `json:"intent"`
	PaymentSource *OrderPaymentSource        `json:"payment_source,omitempty"`
	PurchaseUnits []OrderPurchaseUnitRequest `json:"purchase_units"`
}

type CreateOrderRequest = Order

type OrderOption func(*Order)

func WithOrderShipping() OrderOption {
	return func(order *Order) {
		context := order.paypalExperienceContext()
		context.ShippingPreference = "GET_FROM_FILE"
	}
}

func WithoutOrderShipping() OrderOption {
	return func(order *Order) {
		context := order.paypalExperienceContext()
		context.ShippingPreference = "NO_SHIPPING"
	}
}

func WithOrderImmediatePayment() OrderOption {
	return func(order *Order) {
		context := order.paypalExperienceContext()
		context.PaymentMethodPreference = "IMMEDIATE_PAYMENT_REQUIRED"
	}
}

func WithOrderIntent(intent string) OrderOption {
	return func(order *Order) {
		order.Intent = intent
	}
}

func WithOrderPaypalCountry(countryCode string) OrderOption {
	return func(order *Order) {
		source := order.paypalPaymentSource()
		source.Address = &PostalAddress{CountryCode: countryCode}
	}
}

func (order *Order) paypalPaymentSource() *OrderPaypalPaymentSource {
	if order.PaymentSource == nil {
		order.PaymentSource = &OrderPaymentSource{}
	}
	return &order.PaymentSource.Paypal
}

func (order *Order) paypalExperienceContext() *OrderPaypalExperienceContext {
	source := order.paypalPaymentSource()
	if source.ExperienceContext == nil {
		source.ExperienceContext = &OrderPaypalExperienceContext{}
	}
	return source.ExperienceContext
}

type OrderPaymentSource struct {
	Paypal OrderPaypalPaymentSource `json:"paypal"`
}

type OrderPaypalPaymentSource struct {
	ExperienceContext *OrderPaypalExperienceContext `json:"experience_context,omitempty"`
	Address           *PostalAddress                `json:"address,omitempty"`
	BillingAddress    *PostalAddress                `json:"billing_address,omitempty"`
}

type OrderPaypalExperienceContext struct {
	PaymentMethodPreference string `json:"payment_method_preference,omitempty"`
	LandingPage             string `json:"landing_page,omitempty"`
	ShippingPreference      string `json:"shipping_preference,omitempty"`
	UserAction              string `json:"user_action,omitempty"`
	ReturnURL               string `json:"return_url,omitempty"`
	CancelURL               string `json:"cancel_url,omitempty"`
}

type OrderPurchaseUnitRequest struct {
	InvoiceID string      `json:"invoice_id,omitempty"`
	Amount    OrderAmount `json:"amount"`
	Items     []OrderItem `json:"items,omitempty"`
}

type OrderAmount struct {
	CurrencyCode string                `json:"currency_code"`
	Value        string                `json:"value"`
	Breakdown    *OrderAmountBreakdown `json:"breakdown,omitempty"`
}

type OrderAmountBreakdown struct {
	ItemTotal *Money `json:"item_total,omitempty"`
	Shipping  *Money `json:"shipping,omitempty"`
}

type OrderItem struct {
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	UnitAmount  Money     `json:"unit_amount"`
	Quantity    string    `json:"quantity"`
	Category    string    `json:"category,omitempty"`
	Sku         string    `json:"sku,omitempty"`
	ImageURL    string    `json:"image_url,omitempty"`
	URL         string    `json:"url,omitempty"`
	Upc         *OrderUPC `json:"upc,omitempty"`
}

type OrderUPC struct {
	Type string `json:"type"`
	Code string `json:"code"`
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
			Address PostalAddress `json:"address"`
		} `json:"shipping"`
		Payments struct {
			Captures []struct {
				ID               string `json:"id"`
				Status           string `json:"status"`
				Amount           Money  `json:"amount"`
				SellerProtection struct {
					Status            string   `json:"status"`
					DisputeCategories []string `json:"dispute_categories"`
				} `json:"seller_protection"`
				FinalCapture              bool `json:"final_capture"`
				SellerReceivableBreakdown struct {
					GrossAmount Money `json:"gross_amount"`
					PaypalFee   Money `json:"paypal_fee"`
					NetAmount   Money `json:"net_amount"`
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

func (c *Client) CreateOrder(ctx context.Context, currencyCode, amount string, opts ...OrderOption) (*CreateOrderResponse, error) {
	order := Order{
		Intent: "CAPTURE",
		PurchaseUnits: []OrderPurchaseUnitRequest{
			{
				Amount: OrderAmount{
					CurrencyCode: currencyCode,
					Value:        amount,
				},
			},
		},
	}

	for _, opt := range opts {
		opt(&order)
	}

	response := &CreateOrderResponse{}
	if err := c.api(ctx, http.MethodPost, createOrderEndpoint, order, response); err != nil {
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
	if err := c.api(ctx, http.MethodPost, endpoint, nil, response); err != nil {
		return nil, err
	}
	return response, nil
}
