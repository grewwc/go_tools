package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEvaluateExpression(t *testing.T) {
	tests := []struct {
		name   string
		expr   string
		prec   int
		degree bool
		want   string
	}{
		{name: "nested parentheses", expr: "2*(3+(4-1))", prec: defaultPrec, want: "12"},
		{name: "implicit multiplication", expr: "(1+2)(3+4)", prec: defaultPrec, want: "21"},
		{name: "constants and functions", expr: "2pi + sqrt(81)", prec: defaultPrec, want: "15.2831853071795864"},
		{name: "negative exponent", expr: "2^-2", prec: defaultPrec, want: "0.25"},
		{name: "right associative power", expr: "2^3^2", prec: defaultPrec, want: "512"},
		{name: "unary minus after power", expr: "-2^2", prec: defaultPrec, want: "-4"},
		{name: "degree trig", expr: "sin(30)+cos(60)", prec: defaultPrec, degree: true, want: "1"},
		{name: "min max mix", expr: "max(2,3*4,sqrt(81))", prec: defaultPrec, want: "12"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluateExpression(tt.expr, tt.prec, tt.degree)
			if err != nil {
				t.Fatalf("evaluateExpression(%q) error = %v", tt.expr, err)
			}
			if got != tt.want {
				t.Fatalf("evaluateExpression(%q) = %q, want %q", tt.expr, got, tt.want)
			}
		})
	}
}

func TestParseCLIArgs(t *testing.T) {
	opts, exprArgs, err := parseCLIArgs([]string{"--deg", "--prec=8", "1", "-", "2"})
	if err != nil {
		t.Fatalf("parseCLIArgs() error = %v", err)
	}
	if !opts.degree {
		t.Fatal("expected degree mode to be enabled")
	}
	if opts.prec != 8 {
		t.Fatalf("expected precision 8, got %d", opts.prec)
	}
	if strings.Join(exprArgs, " ") != "1 - 2" {
		t.Fatalf("unexpected expression args: %v", exprArgs)
	}

	opts, exprArgs, err = parseCLIArgs([]string{"--expr=-2^3", "--"})
	if err != nil {
		t.Fatalf("parseCLIArgs() inline expr error = %v", err)
	}
	if opts.expr != "-2^3" {
		t.Fatalf("expected inline expr to be preserved, got %q", opts.expr)
	}
	if len(exprArgs) != 0 {
		t.Fatalf("expected no positional args, got %v", exprArgs)
	}

	_, exprArgs, err = parseCLIArgs([]string{"--", "-2", "+", "5"})
	if err != nil {
		t.Fatalf("parseCLIArgs() double-dash error = %v", err)
	}
	if strings.Join(exprArgs, " ") != "-2 + 5" {
		t.Fatalf("unexpected args after --: %v", exprArgs)
	}
}

func TestResolveExpressionInput(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "expr.txt")
	if err := os.WriteFile(file, []byte("1 + 2\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got, err := resolveExpressionInput("", file, nil, strings.NewReader(""), true)
	if err != nil {
		t.Fatalf("resolveExpressionInput(file) error = %v", err)
	}
	if got != "1 + 2" {
		t.Fatalf("resolveExpressionInput(file) = %q, want %q", got, "1 + 2")
	}

	got, err = resolveExpressionInput("", "", []string{"1", "-", "2"}, strings.NewReader(""), true)
	if err != nil {
		t.Fatalf("resolveExpressionInput(args) error = %v", err)
	}
	if got != "1 - 2" {
		t.Fatalf("resolveExpressionInput(args) = %q, want %q", got, "1 - 2")
	}

	got, err = resolveExpressionInput("", "", nil, strings.NewReader("2(3+4)\n"), false)
	if err != nil {
		t.Fatalf("resolveExpressionInput(stdin) error = %v", err)
	}
	if got != "2(3+4)" {
		t.Fatalf("resolveExpressionInput(stdin) = %q, want %q", got, "2(3+4)")
	}

	if _, err := resolveExpressionInput("", "", nil, strings.NewReader(""), true); err == nil {
		t.Fatal("expected missing expression error")
	}
}
