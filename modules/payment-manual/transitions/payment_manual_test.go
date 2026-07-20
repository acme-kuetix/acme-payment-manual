package transitions

import (
	"context"
	"testing"

	payment "github.com/acme-kuetix/acme-payment/modules/payment/transitions"
)

// TestManualProviderRegistered verifies that blank-importing this
// package (which the test binary does by virtue of being in the same
// package) registered the "manual" provider into acme-payment's
// registry.
func TestManualProviderRegistered(t *testing.T) {
	p, err := payment.GetProvider("manual")
	if err != nil {
		t.Fatalf("manual provider not registered: %v", err)
	}
	if _, ok := p.(*ManualProvider); !ok {
		t.Errorf("expected *ManualProvider, got %T", p)
	}
}

// TestManualProviderAuthorize returns the expected reference + state.
func TestManualProviderAuthorize(t *testing.T) {
	m := &ManualProvider{}
	ctx := map[string]interface{}{
		payment.PaymentContextPayment: map[string]interface{}{"id": "P-1", "amount": 100.0},
		payment.PaymentContextMethod:   map[string]interface{}{"methodType": "cash"},
		payment.PaymentContextAmount:   100.0,
	}
	r := m.Authorize(context.Background(), ctx)
	if ref, _ := r[payment.ProviderResultReference].(string); ref != "MANUAL-AUTH-P-1" {
		t.Errorf("expected MANUAL-AUTH-P-1, got %v", r[payment.ProviderResultReference])
	}
	if state, _ := r[payment.ProviderResultState].(string); state != payment.PaymentStateAuthorized {
		t.Errorf("expected authorized, got %v", r[payment.ProviderResultState])
	}
}

func TestManualProviderCapture(t *testing.T) {
	m := &ManualProvider{}
	ctx := map[string]interface{}{
		payment.PaymentContextPayment: map[string]interface{}{"id": "P-2", "amount": 50.0},
		payment.PaymentContextAmount:  50.0,
	}
	r := m.Capture(context.Background(), ctx)
	if state, _ := r[payment.ProviderResultState].(string); state != payment.PaymentStateCaptured {
		t.Errorf("expected captured, got %v", r[payment.ProviderResultState])
	}
	if ref, _ := r[payment.ProviderResultReference].(string); ref != "MANUAL-CAP-P-2" {
		t.Errorf("expected MANUAL-CAP-P-2, got %v", r[payment.ProviderResultReference])
	}
}

func TestManualProviderRefundFull(t *testing.T) {
	m := &ManualProvider{}
	ctx := map[string]interface{}{
		payment.PaymentContextPayment: map[string]interface{}{"id": "P-3", "amount": 100.0},
		payment.PaymentContextAmount:  100.0,
	}
	r := m.Refund(context.Background(), ctx)
	if state, _ := r[payment.ProviderResultState].(string); state != payment.PaymentStateRefunded {
		t.Errorf("expected refunded for full refund, got %v", r[payment.ProviderResultState])
	}
}

func TestManualProviderRefundPartial(t *testing.T) {
	m := &ManualProvider{}
	ctx := map[string]interface{}{
		payment.PaymentContextPayment: map[string]interface{}{"id": "P-4", "amount": 100.0},
		payment.PaymentContextAmount:  30.0,
	}
	r := m.Refund(context.Background(), ctx)
	if state, _ := r[payment.ProviderResultState].(string); state != payment.PaymentStateCaptured {
		t.Errorf("expected captured for partial refund, got %v", r[payment.ProviderResultState])
	}
}

func TestManualProviderVoid(t *testing.T) {
	m := &ManualProvider{}
	ctx := map[string]interface{}{
		payment.PaymentContextPayment: map[string]interface{}{"id": "P-5", "amount": 75.0},
		payment.PaymentContextAmount:  75.0,
	}
	r := m.Void(context.Background(), ctx)
	if state, _ := r[payment.ProviderResultState].(string); state != payment.PaymentStateVoided {
		t.Errorf("expected voided, got %v", r[payment.ProviderResultState])
	}
}

func TestManualProviderStatus(t *testing.T) {
	m := &ManualProvider{}
	ctx := map[string]interface{}{
		payment.PaymentContextPayment: map[string]interface{}{
			"id":          "P-6",
			"state":       payment.PaymentStateCaptured,
			"providerRef": "MANUAL-CAP-P-6",
		},
		payment.PaymentContextAmount: 100.0,
	}
	r := m.Status(context.Background(), ctx)
	if state, _ := r[payment.ProviderResultState].(string); state != payment.PaymentStateCaptured {
		t.Errorf("expected captured echoed back, got %v", r[payment.ProviderResultState])
	}
	if ref, _ := r[payment.ProviderResultReference].(string); ref != "MANUAL-CAP-P-6" {
		t.Errorf("expected ref echoed back, got %v", r[payment.ProviderResultReference])
	}
}

// TestManualProviderListed ensures the provider appears in the registry listing.
func TestManualProviderListed(t *testing.T) {
	found := false
	for _, c := range payment.ListProviders() {
		if c == "manual" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("manual not in provider list %v", payment.ListProviders())
	}
}
