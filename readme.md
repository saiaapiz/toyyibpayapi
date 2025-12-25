# toyyibpayapi

Simple Go client for ToyyibPay API.

## Installation

go get github.com/saiaapiz/toyyibpayapi

(Replace `github.com/saiaapiz/toyyibpayapi` with the module path where you publish this package.)

## Usage

Import and create a client with your secret key:

```go
import "github.com/saiaapiz/toyyibpayapi"

api := toyyibpayapi.New("YOUR_SECRET_KEY", true) // debug = true (optional)
```

Get category details:

```go
cat, err := api.Category("CATEGORY_CODE")
if err != nil { /* handle error */ }
fmt.Println(cat.Name, cat.Active())
```

Create a bill (amount in sen — 1000 = RM10.00):

```go
billcode, err := api.Create(
    "CATEGORY_CODE",
    "Title",
    "Description",
    1000,
    toyyibpayapi.WithWebhook("ref-001", "https://example.com/webhook", "https://example.com/return"),
    toyyibpayapi.WithPayerInfo("Name", "0123456789", "user@example.com"),
)
if err != nil { /* handle error */ }
fmt.Println("https://toyyibpay.com/" + billcode)
```

Deactivate a bill:

```go
if err := api.Deactive(billcode); err != nil { /* handle error */ }
```

## Notes

- Amounts are provided in sen (1 MYR = 100 sen).
- The client normalizes some non-standard ToyyibPay responses; check errors for backend messages.
- Use debug mode to print raw responses.

## Tests

See example tests in repository. Provide your real secret and category code to run integration tests.

## License

MIT