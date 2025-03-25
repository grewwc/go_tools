package test

import (
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/grewwc/go_tools/src/numW"
	"github.com/grewwc/go_tools/src/strW"
)

func TestPlus(t *testing.T) {
	for i := 0; i < 1000; i++ {
		arr := numW.RandInt(-1000, 1000, 2)
		a, b := arr[0], arr[1]
		res := strW.Plus(strconv.Itoa(a), strconv.Itoa(b))
		if strconv.Itoa(a+b) != res {
			t.Errorf("Plus(%d, %d) = %d, want |%s|", a, b, a+b, res)
		}
	}
}

func TestMinus(t *testing.T) {
	for i := 0; i < 1000; i++ {
		arr := numW.RandInt(-1000, 1000, 2)
		a, b := arr[0], arr[1]
		ba := big.NewInt(int64(a))
		bb := big.NewInt(int64(b))
		expect := ba.Sub(ba, bb)
		res := strW.Minus(strconv.Itoa(a), strconv.Itoa(b))
		if expect.String() != res {
			t.Errorf("Minus(%d, %d) = %s, want |%s|", a, b, expect.String(), res)
		}
	}
}

func TestMul(t *testing.T) {
	for i := 0; i < 10000; i++ {
		arr := numW.RandInt(-1000, 1000, 2)
		a, b := arr[0], arr[1]
		ba := big.NewInt(int64(a))
		bb := big.NewInt(int64(b))
		expect := ba.Mul(ba, bb)
		res := strW.Mul(strconv.Itoa(a), strconv.Itoa(b))
		if expect.String() != res {
			t.Errorf("Mul(%d, %d) = %s, want |%s|", a, b, expect.String(), res)
		}
	}

}

func TestDiv(t *testing.T) {
	for i := 0; i < 1000; i++ {
		arr := numW.RandInt(-100, 100, 2)
		a, b := arr[0], arr[1]
		if b == 0 {
			continue
		}
		res := strW.Div(strconv.Itoa(a), strconv.Itoa(b), 30)
		recoverBack := strW.Mul(res, strconv.Itoa(b))
		diff := strW.Minus(strconv.Itoa(a), recoverBack)
		if strings.Count(diff, "0") < 10 && diff != "0" {
			t.Errorf("Div(%d, %d, 30) = %s, recoverBack: |%s|, diff |%s| ", a, b, res, recoverBack, diff)
		}
	}

}
