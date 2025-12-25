package toyyibpayapi

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-faster/errors"
	"resty.dev/v3"
)

type category struct {
	Id          string `json:"-"`
	Name        string `json:"CategoryName"`
	Description string `json:"categoryDescription"`
	Status      string `json:"categoryStatus"`
}

func (c category) Active() bool {
	return c.Status == "1"
}

type client struct {
	secret string
	resty  *resty.Client
}

func New(secret string, debug ...bool) *client {
	resty := resty.New()
	if len(debug) > 0 && debug[0] {
		resty.SetDebug(true)
	}

	return &client{
		secret: secret,
		resty: resty.
			SetHeader("User-Agent", "toyyibpay-go-client/0.1").
			SetFormData(map[string]string{"userSecretKey": secret, "secretKey": secret}).
			SetBaseURL("https://toyyibpay.com/index.php/api/"),
	}
}

// https://toyyibpay.com/apireference/#gc
func (c *client) Category(code string) (*category, error) {
	var category *category

	r := c.resty.R()
	r.ForceResponseContentType = "application/json"

	resp, err := r.
		SetFormData(map[string]string{"categoryCode": code}).
		SetResult(&category).
		Post("getCategoryDetails")

	if err != nil {
		return nil, errors.Wrap(err, "resty-request")
	}

	if category == nil {
		return nil, errors.New("invalid-response")
	}

	// toyyibpay returns 200 OK even for invalid category codes
	if category.Name == "" && category.Description == "" && category.Status == "" {
		return nil, errors.New("invalid category code")
	}

	_ = resp

	category.Id = code
	return category, nil
}

// https://toyyibpay.com/apireference/#ib
func (c *client) Deactive(billcode string) error {
	var data struct {
		Status string `json:"status"`
		Result string `json:"result"`
	}

	r := c.resty.R()
	r.ForceResponseContentType = "application/json"

	resp, err := r.
		SetFormData(map[string]string{"billCode": billcode}).
		SetResult(&data).
		Post("inactiveBill")

	if err != nil {
		return errors.Wrap(err, "resty-request")
	}

	if data.Status != "success" {
		return fmt.Errorf("toyyibpay-backend: %s", data.Result)
	}

	_ = resp

	return nil
}

type billRequest struct {
	Category string `json:"categoryCode"`

	Amount      int    `json:"billAmount"`
	Title       string `json:"billName"`                  //* Max 30 alphanumeric characters, space and '_' only
	Description string `json:"billDescription,omitempty"` //* Max 30 alphanumeric characters, space and '_' only

	WebhookReference string `json:"billExternalReferenceNo"`
	WebhookURL       string `json:"billCallbackUrl"`
	ReturnURL        string `json:"billReturnUrl,omitempty"`

	CustomPayerInfo int    `json:"billPayorInfo"`
	Email           string `json:"billEmail"`
	Phone           string `json:"billPhone"`
	CustomerName    string `json:"billTo,omitempty"`

	DynamicPrice     int    `json:"billPriceSetting"`
	ChargeToCustomer int    `json:"billChargeToCustomer"`
	ChargeToPrepaid  int    `json:"billChargeToPrepaid,omitempty"`
	EmailContent     string `json:"billContentEmail,omitempty"`
	ExpiryDate       string `json:"billExpiryDate,omitempty"` // Example value: "17-12-2020 17:00:00"
	ExpiryDays       int    `json:"billExpiryDays,omitempty"` // Default 1 days. Ranged between 1 to 100 days, billExpiryDate takes precedence if both are set.

	PaymentChannel     int `json:"billPaymentChannel"` // 0: FPX, 1: Credit/Debit Card, 2: Both
	EnableCorporateFPX int `json:"enableFPXB2B,omitempty"`
	ChargeCorporateFPX int `json:"chargeFPXB2B,omitempty"`

	SplitPayment     int    `json:"billSplitPayment,omitempty"`
	SplitPaymentArgs string `json:"billSplitPaymentArgs,omitempty"`
}

func (br billRequest) IntoForms() map[string]string {
	var t map[string]any
	b, _ := json.Marshal(br)
	_ = json.Unmarshal(b, &t)

	m := make(map[string]string, len(t))
	for k, v := range t {
		m[k] = fmt.Sprint(v)
	}

	return m
}

type billOptions func(*billRequest)

func WithPayerInfo(Name, Phone, Email string) billOptions {
	return func(br *billRequest) {
		br.CustomPayerInfo = 1
		br.CustomerName = Name
		br.Phone = Phone
		br.Email = Email
	}
}

func WithWebhook(Reference, URL string, ReturnUrl ...string) billOptions {
	return func(br *billRequest) {
		if len(ReturnUrl) > 0 {
			br.ReturnURL = ReturnUrl[0]
		}
		br.WebhookReference = Reference
		br.WebhookURL = URL
	}
}

// https://toyyibpay.com/apireference/#cb
func (c *client) Create(CategoryCode, Title, Description string, AmountInSen int, options ...billOptions) (string, error) {
	pdata := &billRequest{
		Category:           CategoryCode,
		Title:              Title,
		Description:        Description,
		Amount:             AmountInSen,
		EnableCorporateFPX: 1,
		ChargeCorporateFPX: 1,
	}

	for _, opt := range options {
		opt(pdata)
	}

	r := c.resty.R()
	r.ForceResponseContentType = "application/json"
	r.DoNotParseResponse = true

	resp, err := r.
		SetFormData(pdata.IntoForms()).
		Post("createBill")

	if err != nil {
		return "", errors.Wrap(err, "resty-request")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "read-body")
	}
	resp.Body.Close()

	if c.resty.IsDebug() {
		fmt.Println("Responses Body:", string(body))
	}

	// Toyyibpay sometimes returns non-standard JSON responses.
	// Try to decode expected response first, then fallback to error response.

	var billResponses []struct {
		Billcode string `json:"billCode"`
	}

	var msgResponses struct {
		Status  string `json:"status,omitempty"`
		Message string `json:"msg,omitempty"`
	}

	if err := json.Unmarshal(body, &billResponses); err != nil {
		if err := json.Unmarshal(body, &msgResponses); err != nil {
			return "", errors.Wrap(err, "decode-response")
		}

		if msgResponses.Status != "success" {
			return "", fmt.Errorf("toyyibpay-backend: %s", msgResponses.Message)
		}
	}

	_ = resp

	return billResponses[0].Billcode, nil
}
