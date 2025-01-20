package stringsW

import (
	"math/big"
	"strconv"
	"testing"

	"github.com/grewwc/go_tools/src/randW"
	"github.com/grewwc/go_tools/src/stringsW"
)

func TestPlus(t *testing.T) {
	for i := 0; i < 100; i++ {
		arr := randW.RandInt(-1000000, 1000000, 2)
		a, b := arr[0], arr[1]
		res := stringsW.Plus(strconv.Itoa(a), strconv.Itoa(b))
		if strconv.Itoa(a+b) != res {
			t.Errorf("Plus(%d, %d) = %d, want |%s|", a, b, a+b, res)
		}
	}
}

func TestMinus(t *testing.T) {
	for i := 0; i < 100; i++ {
		arr := randW.RandInt(-10000000, 1000000000, 2)
		a, b := arr[0], arr[1]
		ba := big.NewInt(int64(a))
		bb := big.NewInt(int64(b))
		expect := ba.Sub(ba, bb)
		res := stringsW.Minus(strconv.Itoa(a), strconv.Itoa(b))
		if expect.String() != res {
			t.Errorf("Minus(%d, %d) = %s, want |%s|", a, b, expect.String(), res)
		}
	}
}

func TestMul(t *testing.T) {
	for i := 0; i < 100; i++ {
		arr := randW.RandInt(-10000000, 1000000000, 2)
		a, b := arr[0], arr[1]
		ba := big.NewInt(int64(a))
		bb := big.NewInt(int64(b))
		expect := ba.Mul(ba, bb)
		res := stringsW.Mul(strconv.Itoa(a), strconv.Itoa(b))
		if expect.String() != res {
			t.Errorf("Mul(%d, %d) = %s, want |%s|", a, b, expect.String(), res)
		}
	}

}
