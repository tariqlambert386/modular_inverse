package modular_inverse

import (
	"errors"
	"math/big"
	"testing"
)

// bigFromInt64 is a small helper so tests read cleanly.
func bigFromInt64(v int64) *big.Int {
	return big.NewInt(v)
}

// mustBig parses a decimal string into a big.Int, failing the test on error.
func mustBig(t *testing.T, s string) *big.Int {
	t.Helper()
	v, ok := new(big.Int).SetString(s, 10)
	if !ok {
		t.Fatalf("invalid big.Int literal %q", s)
	}
	return v
}

func TestModInverseBasic(t *testing.T) {
	cases := []struct {
		a, m, want int64
	}{
		{3, 11, 4},   // 3*4 = 12 ≡ 1
		{10, 17, 12}, // 10*12 = 120 ≡ 1
		{1, 5, 1},
		{4, 5, 4},
	}
	for _, tc := range cases {
		got, err := ModInverse(bigFromInt64(tc.a), bigFromInt64(tc.m))
		if err != nil {
			t.Fatalf("ModInverse(%d, %d) unexpected error: %v", tc.a, tc.m, err)
		}
		if got.Cmp(bigFromInt64(tc.want)) != 0 {
			t.Errorf("ModInverse(%d, %d) = %v, want %d", tc.a, tc.m, got, tc.want)
		}
	}
}

func TestModInverseNegativeA(t *testing.T) {
	// -3 ≡ 8 (mod 11), and 8*7 = 56 ≡ 1 (mod 11).
	got, err := ModInverse(bigFromInt64(-3), bigFromInt64(11))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := bigFromInt64(7)
	if got.Cmp(want) != 0 {
		t.Errorf("ModInverse(-3, 11) = %v, want %v", got, want)
	}
}

func TestModInverseNoInverse(t *testing.T) {
	// gcd(6, 9) = 3, so 6 has no inverse mod 9.
	got, err := ModInverse(bigFromInt64(6), bigFromInt64(9))
	if !errors.Is(err, ErrNoInverse) {
		t.Fatalf("ModInverse(6, 9) error = %v, want ErrNoInverse", err)
	}
	if got != nil {
		t.Errorf("ModInverse(6, 9) = %v, want nil", got)
	}
}

func TestModInverseInvalidModulus(t *testing.T) {
	for _, m := range []int64{0, -5, 1} {
		got, err := ModInverse(bigFromInt64(3), bigFromInt64(m))
		if err == nil {
			t.Errorf("ModInverse(3, %d) expected error, got %v", m, got)
		}
		if got != nil {
			t.Errorf("ModInverse(3, %d) = %v, want nil", m, got)
		}
	}
}

func TestModInverseLarge(t *testing.T) {
	// A large prime and a value coprime to it.
	m := mustBig(t, "170141183460469231731687303715884105727") // 2^127 - 1
	a := mustBig(t, "12345678901234567890123456789012345678")
	inv, err := ModInverse(a, m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify a*inv ≡ 1 (mod m).
	prod := new(big.Int).Mul(a, inv)
	prod.Mod(prod, m)
	if prod.Cmp(big.NewInt(1)) != 0 {
		t.Errorf("a*inv mod m = %v, want 1", prod)
	}
}

func TestCRTBasic(t *testing.T) {
	// x ≡ 2 (mod 3), x ≡ 3 (mod 5), x ≡ 2 (mod 7)
	// Solution: x = 23 mod 105.
	residues := []*big.Int{bigFromInt64(2), bigFromInt64(3), bigFromInt64(2)}
	moduli := []*big.Int{bigFromInt64(3), bigFromInt64(5), bigFromInt64(7)}
	got, err := CRT(residues, moduli)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := bigFromInt64(23)
	if got.Cmp(want) != 0 {
		t.Errorf("CRT = %v, want %v", got, want)
	}
}

func TestCRTSingleCongruence(t *testing.T) {
	residues := []*big.Int{bigFromInt64(4)}
	moduli := []*big.Int{bigFromInt64(9)}
	got, err := CRT(residues, moduli)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Cmp(bigFromInt64(4)) != 0 {
		t.Errorf("CRT = %v, want 4", got)
	}
}

func TestCRTResidueAlreadyReduced(t *testing.T) {
	// Residues larger than their moduli should be reduced.
	residues := []*big.Int{bigFromInt64(8), bigFromInt64(3)}
	moduli := []*big.Int{bigFromInt64(5), bigFromInt64(7)}
	// 8 ≡ 3 (mod 5), so x ≡ 3 (mod 5) and x ≡ 3 (mod 7) => x = 3 mod 35.
	got, err := CRT(residues, moduli)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Cmp(bigFromInt64(3)) != 0 {
		t.Errorf("CRT = %v, want 3", got)
	}
}

func TestCRTNoSolution(t *testing.T) {
	// x ≡ 1 (mod 2), x ≡ 0 (mod 4) has no solution, but 2 and 4 are not
	// coprime, so we expect ErrNotCoprime rather than ErrNoSolution.
	residues := []*big.Int{bigFromInt64(1), bigFromInt64(0)}
	moduli := []*big.Int{bigFromInt64(2), bigFromInt64(4)}
	_, err := CRT(residues, moduli)
	if !errors.Is(err, ErrNotCoprime) {
		t.Fatalf("CRT error = %v, want ErrNotCoprime", err)
	}
}

func TestCRTNotCoprime(t *testing.T) {
	residues := []*big.Int{bigFromInt64(1), bigFromInt64(2)}
	moduli := []*big.Int{bigFromInt64(6), bigFromInt64(9)}
	_, err := CRT(residues, moduli)
	if !errors.Is(err, ErrNotCoprime) {
		t.Fatalf("CRT error = %v, want ErrNotCoprime", err)
	}
}

func TestCRTLengthMismatch(t *testing.T) {
	residues := []*big.Int{bigFromInt64(1)}
	moduli := []*big.Int{bigFromInt64(3), bigFromInt64(4)}
	_, err := CRT(residues, moduli)
	if err == nil {
		t.Fatal("expected error for mismatched lengths")
	}
}

func TestCRTEmpty(t *testing.T) {
	_, err := CRT(nil, nil)
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestCRTInvalidModulus(t *testing.T) {
	for _, m := range []int64{0, -3, 1} {
		residues := []*big.Int{bigFromInt64(1)}
		moduli := []*big.Int{bigFromInt64(m)}
		_, err := CRT(residues, moduli)
		if err == nil {
			t.Errorf("CRT with modulus %d expected error", m)
		}
	}
}

func TestCRTLarge(t *testing.T) {
	// Three large pairwise-coprime moduli.
	m1 := mustBig(t, "1000000007")
	m2 := mustBig(t, "1000000009")
	m3 := mustBig(t, "1000000021")
	residues := []*big.Int{bigFromInt64(1), bigFromInt64(2), bigFromInt64(3)}
	moduli := []*big.Int{m1, m2, m3}
	got, err := CRT(residues, moduli)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify each congruence.
	for i, r := range residues {
		check := new(big.Int).Mod(got, moduli[i])
		if check.Cmp(r) != 0 {
			t.Errorf("x mod m[%d] = %v, want %v", i, check, r)
		}
	}
}
