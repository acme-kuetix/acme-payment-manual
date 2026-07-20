# acme-payment-manual

Reference payment provider adapter for `acme-payment`. Implements the `Provider` interface with no HTTP — every method returns a deterministic reference and the expected next state. Used for:

- Cash payments
- Cheque payments recorded after receipt
- Bank transfers reconciled manually
- Testing and local development without a payment gateway

## How it registers

Blank-import this package to activate the "manual" provider:

```go
import _ "github.com/acme-kuetix/acme-payment-manual/modules/payment-manual/transitions"
```

The package's `init()` calls `payment.RegisterProvider("manual", &ManualProvider{})`. After that, `Payment.ProviderID = "manual"` is a valid provider code.

## Pattern for future providers

This package is the template. To add a new provider (e.g. Stripe):

1. Create `acme-payment-stripe` with the same structure
2. Implement all 5 `Provider` methods (Authorize, Capture, Refund, Void, Status)
3. Register at `init()` time under a code (e.g. "stripe")
4. Blank-import in `erp-app/modules/modules.go` to activate

The `acme-payment` core package never imports a provider — providers are discovered at runtime via the registry.
