package test

import (
	"fmt"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/grewwc/go_tools/src/algoW"
	"github.com/grewwc/go_tools/src/strW"
)

func TestPlus(t *testing.T) {
	for i := 0; i < 1000; i++ {
		arr := algoW.RandInt(-1000, 1000, 2)
		a, b := arr[0], arr[1]
		res := strW.Plus(strconv.Itoa(a), strconv.Itoa(b))
		if strconv.Itoa(a+b) != res {
			t.Errorf("Plus(%d, %d) = %d, want |%s|", a, b, a+b, res)
		}
	}
}

func TestMinus(t *testing.T) {
	for i := 0; i < 1000; i++ {
		arr := algoW.RandInt(-1000, 1000, 2)
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
		arr := algoW.RandInt(-1000, 1000, 2)
		a, b := arr[0], arr[1]
		ba := big.NewInt(int64(a))
		bb := big.NewInt(int64(b))
		expect := ba.Mul(ba, bb)
		res := strW.Mul(strconv.Itoa(a), strconv.Itoa(b))
		if expect.String() != res {
			t.Errorf("Mul(%d, %d) = %s, want |%s|", a, b, res, expect.String())
		}
	}

}

func genRandomNumber(n int) string {
	rand.Seed(time.Now().UnixNano())

	// 定义最终结果字符串
	var result string

	// 循环生成随机整数并拼接到结果字符串中
	for len(result) < n {
		// 生成一个随机整数
		randomInt := rand.Intn(10) // 生成 0 到 9 之间的随机整数
		// 将整数转换为字符串并追加到结果字符串中
		result += fmt.Sprintf("%d", randomInt)
	}

	return result
}

func BenchmarkMul(b *testing.B) {
	n := 5000
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		s1 := genRandomNumber(n)
		s2 := genRandomNumber(n)
		b.StartTimer()
		strW.Mul(s1, s2)
	}
}

func TestDiv(t *testing.T) {
	for i := 0; i < 1000; i++ {
		arr := algoW.RandInt(-100, 100, 2)
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

func BenchmarkDiv(b *testing.B) {
	n := 5000
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		s1 := genRandomNumber(n)
		s2 := genRandomNumber(n)
		b.StartTimer()
		strW.Div(s1, s2, 100)
	}
}
