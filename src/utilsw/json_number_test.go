package utilsw

import (
	"encoding/json"
	"math"
	"testing"
)

func TestGetTJsonNumberConversion(t *testing.T) {
	j, err := NewJsonFromString(`{"i":123,"f":1.25,"s":456,"bad":1.2,"big":200,"neg":-1}`)
	if err != nil {
		t.Fatalf("NewJsonFromString err: %v", err)
	}

	if _, ok := j.Get("i").(json.Number); !ok {
		t.Fatalf("expected json.Number, got %T", j.Get("i"))
	}

	if got := getT[int](j, "i"); got != 123 {
		t.Fatalf("getT[int] = %v", got)
	}
	if got := getT[int8](j, "i"); got != 123 {
		t.Fatalf("getT[int8] = %v", got)
	}
	if got := getT[uint16](j, "i"); got != 123 {
		t.Fatalf("getT[uint16] = %v", got)
	}

	if got := getT[float64](j, "f"); math.Abs(got-1.25) > 1e-9 {
		t.Fatalf("getT[float64] = %v", got)
	}
	if got := getT[float32](j, "f"); math.Abs(float64(got)-1.25) > 1e-6 {
		t.Fatalf("getT[float32] = %v", got)
	}

	if got := getT[string](j, "s"); got != "456" {
		t.Fatalf("getT[string] = %q", got)
	}

	if got := getT[int](j, "bad"); got != 0 {
		t.Fatalf("getT[int] bad = %v", got)
	}
	if got := getT[int8](j, "big"); got != 0 {
		t.Fatalf("getT[int8] big = %v", got)
	}
	if got := getT[uint](j, "neg"); got != 0 {
		t.Fatalf("getT[uint] neg = %v", got)
	}
}

