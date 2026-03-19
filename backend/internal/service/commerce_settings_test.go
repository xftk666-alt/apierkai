package service

import "testing"

func TestNormalizeCommercePurchaseMode(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty defaults to iframe", input: "", want: CommercePurchaseModeIframe},
		{name: "whitespace defaults to iframe", input: "   ", want: CommercePurchaseModeIframe},
		{name: "native preserved", input: "native", want: CommercePurchaseModeNative},
		{name: "native normalized", input: " Native ", want: CommercePurchaseModeNative},
		{name: "unknown falls back to iframe", input: "legacy", want: CommercePurchaseModeIframe},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := NormalizeCommercePurchaseMode(tc.input); got != tc.want {
				t.Fatalf("NormalizeCommercePurchaseMode(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestNormalizeCommercePaymentProviderSettings(t *testing.T) {
	t.Parallel()

	items, err := NormalizeCommercePaymentProviderSettings([]CommercePaymentProviderSetting{
		{
			Code:        " WeChat_Pay ",
			Name:        "WeChat Pay",
			CheckoutURL: "https://pay.example.com/checkout?order_no={order_no}&amount={amount}",
			Enabled:     true,
		},
	})
	if err != nil {
		t.Fatalf("NormalizeCommercePaymentProviderSettings() unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("NormalizeCommercePaymentProviderSettings() len = %d, want 1", len(items))
	}
	if items[0].Code != "wechat_pay" {
		t.Fatalf("NormalizeCommercePaymentProviderSettings() code = %q, want %q", items[0].Code, "wechat_pay")
	}
	if items[0].SortOrder != 100 {
		t.Fatalf("NormalizeCommercePaymentProviderSettings() sort_order = %d, want 100", items[0].SortOrder)
	}
}

func TestBuildCommerceCheckoutURL(t *testing.T) {
	t.Parallel()

	url, err := BuildCommerceCheckoutURL(
		"https://pay.example.com/checkout?order_no={order_no}&amount={amount}&provider={payment_provider}",
		CommerceCheckoutURLTemplateInput{
			OrderID:         10,
			OrderNo:         "CO20250101000000TEST",
			UserID:          20,
			ProductID:       30,
			PriceID:         40,
			Amount:          68,
			Currency:        "cny",
			PaymentProvider: "wechat_pay",
		},
	)
	if err != nil {
		t.Fatalf("BuildCommerceCheckoutURL() unexpected error: %v", err)
	}

	want := "https://pay.example.com/checkout?order_no=CO20250101000000TEST&amount=68.00&provider=wechat_pay"
	if url != want {
		t.Fatalf("BuildCommerceCheckoutURL() = %q, want %q", url, want)
	}
}
