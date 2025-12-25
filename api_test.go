package toyyibpayapi

import (
	"fmt"
	"testing"

	"github.com/go-faster/errors"
)

// Obtain value below from https://toyyibpay.com/
const (
	secret       = "e07741af-19a6-438a-b869-b3dfe7e6d3b1"
	categoryCode = "XGsIFKnx"
)

func TestCategory(t *testing.T) {
	api := New(secret)
	category, err := api.Category(categoryCode)
	if err != nil {
		t.Fatal(errors.Wrap(err, "get-category"))
	}
	t.Log(category)
}

func TestBilling(t *testing.T) {
	api := New(secret, true)
	billcode, err := api.Create(categoryCode, t.Name(), "Billing Test", 1000,
		WithWebhook("ref-001", "https://your-webhook-url.com", "https://example.com"),
		WithPayerInfo("Test User", "012345789", "example@example.com"),
	)

	if err != nil {
		t.Fatal(errors.Wrap(err, "create-bill"))
	}

	fmt.Println("Payment Url: https://toyyibpay.com/" + billcode)

	err = api.Deactive(billcode)
	if err != nil {
		t.Fatal(errors.Wrap(err, "deactive-bill"))
	}

	t.Log(billcode)
}
