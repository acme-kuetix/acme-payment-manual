// Package transitions implements the "manual" payment provider — a
// no-HTTP adapter that records payment intent without contacting any
// external gateway. It's the reference implementation showing how
// external providers plug into acme-payment's Provider registry.
//
// Register by blank-importing this package:
//
//	import _ "github.com/acme-kuetix/acme-payment-manual/modules/payment-manual/transitions"
//
// The init() function calls transitions.RegisterProvider("manual", &ManualProvider{}),
// making the "manual" provider code available for use in Payment.ProviderID.
package transitions

import (
	"context"
	"fmt"

	coreutils "github.com/acme-kuetix/acme-std-core/modules/core/utils/transitions"
	payment "github.com/acme-kuetix/acme-payment/modules/payment/transitions"
)

// ManualProvider implements payment.Provider for offline/manual
// payments: cash, cheque, bank transfers recorded after the fact, etc.
// Every method returns a deterministic reference and the expected
// next state — no HTTP, no retries, no external state to poll.
type ManualProvider struct{}

// code is the provider code this adapter registers under.
const code = "manual"

// paymentID extracts the payment id from the context map.
func paymentID(ctx map[string]interface{}) string {
	if p, ok := ctx[payment.PaymentContextPayment].(map[string]interface{}); ok {
		if id, ok := p["id"].(string); ok {
			return id
		}
	}
	return ""
}

// paymentAmount extracts the payment amount from the context map.
func paymentAmount(ctx map[string]interface{}) float64 {
	if p, ok := ctx[payment.PaymentContextPayment].(map[string]interface{}); ok {
		return coreutils.ToFloatVal(p["amount"])
	}
	return 0
}

// paymentState extracts the payment state from the context map.
func paymentState(ctx map[string]interface{}) string {
	if p, ok := ctx[payment.PaymentContextPayment].(map[string]interface{}); ok {
		if s, ok := p["state"].(string); ok {
			return s
		}
	}
	return ""
}

// paymentProviderRef extracts the providerRef from the context map.
func paymentProviderRef(ctx map[string]interface{}) string {
	if p, ok := ctx[payment.PaymentContextPayment].(map[string]interface{}); ok {
		if s, ok := p["providerRef"].(string); ok {
			return s
		}
	}
	return ""
}

// ctxAmount extracts the operation amount from the context map.
func ctxAmount(ctx map[string]interface{}) float64 {
	return coreutils.ToFloatVal(ctx[payment.PaymentContextAmount])
}

// Authorize records intent. For manual payments there's no
// reservation step — the money either arrived or it didn't — so we
// move straight to authorized with a synthetic reference.
func (m *ManualProvider) Authorize(_ context.Context, ctx map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		payment.ProviderResultReference: fmt.Sprintf("MANUAL-AUTH-%s", paymentID(ctx)),
		payment.ProviderResultState:      payment.PaymentStateAuthorized,
		payment.ProviderResultRawResponse: map[string]interface{}{
			"mode":   "manual",
			"action": "authorize",
		},
	}
}

// Capture records settlement. The money has been received; the
// reference is updated to reflect the capture.
func (m *ManualProvider) Capture(_ context.Context, ctx map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		payment.ProviderResultReference: fmt.Sprintf("MANUAL-CAP-%s", paymentID(ctx)),
		payment.ProviderResultState:      payment.PaymentStateCaptured,
		payment.ProviderResultRawResponse: map[string]interface{}{
			"mode":   "manual",
			"action": "capture",
		},
	}
}

// Refund records a manual reversal (e.g. a cheque returned, a bank
// transfer reversed). Full refunds return State=refunded; partial
// refunds return State=captured and the core package records the
// accumulated refunded amount.
func (m *ManualProvider) Refund(_ context.Context, ctx map[string]interface{}) map[string]interface{} {
	amt := ctxAmount(ctx)
	original := paymentAmount(ctx)
	isFull := amt >= original-1e-6
	state := payment.PaymentStateRefunded
	if !isFull {
		state = payment.PaymentStateCaptured
	}
	return map[string]interface{}{
		payment.ProviderResultReference: fmt.Sprintf("MANUAL-REF-%s", paymentID(ctx)),
		payment.ProviderResultState:      state,
		payment.ProviderResultRawResponse: map[string]interface{}{
			"mode":   "manual",
			"action": "refund",
			"amount": amt,
			"full":   isFull,
		},
	}
}

// Void cancels an authorization. For manual payments this is just a
// bookkeeping flip — no external party to notify.
func (m *ManualProvider) Void(_ context.Context, ctx map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		payment.ProviderResultReference: fmt.Sprintf("MANUAL-VOID-%s", paymentID(ctx)),
		payment.ProviderResultState:      payment.PaymentStateVoided,
		payment.ProviderResultRawResponse: map[string]interface{}{
			"mode":   "manual",
			"action": "void",
		},
	}
}

// Status returns the current locally-recorded state. Manual payments
// have no async settlement to poll.
func (m *ManualProvider) Status(_ context.Context, ctx map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		payment.ProviderResultReference: paymentProviderRef(ctx),
		payment.ProviderResultState:      paymentState(ctx),
		payment.ProviderResultRawResponse: map[string]interface{}{
			"mode":   "manual",
			"action": "status",
		},
	}
}

func init() {
	payment.RegisterProvider(code, &ManualProvider{})
}
