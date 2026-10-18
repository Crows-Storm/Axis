package exchange

import "testing"

func TestIntervalValid(t *testing.T) {
	if !OneHour.Valid() {
		t.Fatal("expected 1h to be valid")
	}
	if Interval("bad").Valid() {
		t.Fatal("expected invalid interval")
	}
	if Interval("").Valid() {
		t.Fatal("expected empty interval to be invalid")
	}
}

func TestParsePlatform(t *testing.T) {
	p, ok := ParsePlatform("binance")
	if !ok || p != Binance {
		t.Fatalf("binance: got %q %v", p, ok)
	}
	p, ok = ParsePlatform("OKX")
	if !ok || p != Okx {
		t.Fatalf("okx: got %q %v", p, ok)
	}
	if _, ok := ParsePlatform("unknown"); ok {
		t.Fatal("expected unknown platform to fail")
	}
}
