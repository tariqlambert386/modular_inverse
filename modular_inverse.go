// Package modular_inverse provides modular multiplicative inverses and
// Chinese remainder theorem solutions over big integers.
//
// The implementation is deliberately small and standard-library only. All
// arithmetic uses math/big so that inputs are not limited by the machine word
// size. The package makes one clear decision about the Chinese remainder
// theorem: moduli must be pairwise coprime. If they are not, the solution is
// not unique modulo the product, and silently returning one representative
// would be misleading. Callers who need non-coprime moduli should combine the
// congruences themselves with the extended Euclidean algorithm.
package modular_inverse

import (
	"errors"
	"math/big"
)

// ErrNoInverse is returned by ModInverse when the inverse does not exist.
// It is a sentinel so callers can test for it with errors.Is.
var ErrNoInverse = errors.New("modular_inverse: no inverse exists")

// ErrNoSolution is returned by CRT when the system has no solution.
var ErrNoSolution = errors.New("modular_inverse: no solution")

// ErrNotCoprime is returned by CRT when the moduli are not pairwise coprime.
var ErrNotCoprime = errors.New("modular_inverse: moduli are not pairwise coprime")

// ModInverse returns the multiplicative inverse of a modulo m.
//
// The inverse is the unique x in [0, m) such that a*x ≡ 1 (mod m).
// It exists if and only if gcd(a, m) == 1. If no inverse exists, the returned
// value is nil and the error is ErrNoInverse.
//
// m must be greater than 1. A modulus of 1 would make every integer congruent
// to 0, and the inverse would be ill-defined.
func ModInverse(a, m *big.Int) (*big.Int, error) {
	if m == nil || m.Sign() <= 0 {
		return nil, errors.New("modular_inverse: modulus must be positive")
	}
	if m.Cmp(big.NewInt(1)) == 0 {
		return nil, errors.New("modular_inverse: modulus must be greater than 1")
	}

	// big.Int.ModInverse returns nil when gcd(a, m) != 1.
	inv := new(big.Int).ModInverse(a, m)
	if inv == nil {
		return nil, ErrNoInverse
	}
	return inv, nil
}

// CRT solves the system of congruences
//
//	x ≡ residues[i] (mod moduli[i])
//
// for i in [0, n). The moduli must be pairwise coprime and greater than 1.
// The returned solution is the unique value in [0, M) where M is the product
// of the moduli. If the system has no solution, CRT returns ErrNoSolution.
// If the moduli are not pairwise coprime, CRT returns ErrNotCoprime.
//
// The pairwise-coprime restriction is a deliberate choice. For non-coprime
// moduli a solution may exist or not, and when it exists it is only unique
// modulo the least common multiple. Supporting that would require returning
// both a value and a modulus, which complicates the common case. Callers who
// need it can reduce the system themselves.
func CRT(residues, moduli []*big.Int) (*big.Int, error) {
	if len(residues) != len(moduli) {
		return nil, errors.New("modular_inverse: residues and moduli must have the same length")
	}
	if len(residues) == 0 {
		return nil, errors.New("modular_inverse: at least one congruence is required")
	}

	// Validate moduli and check pairwise coprimality up front so the
	// incremental merge below can assume it.
	for i, m := range moduli {
		if m == nil || m.Sign() <= 0 {
			return nil, errors.New("modular_inverse: moduli must be positive")
		}
		if m.Cmp(big.NewInt(1)) == 0 {
			return nil, errors.New("modular_inverse: moduli must be greater than 1")
		}
		for j := 0; j < i; j++ {
			g := new(big.Int).GCD(nil, nil, moduli[i], moduli[j])
			if g.Cmp(big.NewInt(1)) != 0 {
				return nil, ErrNotCoprime
			}
		}
	}

	// Incremental CRT: combine two congruences at a time.
	//
	// Given x ≡ r1 (mod m1) and x ≡ r2 (mod m2) with gcd(m1, m2) == 1,
	// the combined congruence is x ≡ r (mod m1*m2) where
	//
	//	r = r1 + m1 * ((r2 - r1) * inv(m1, m2) mod m2)
	//
	// and inv(m1, m2) is the inverse of m1 modulo m2.
	result := new(big.Int).Set(residues[0])
	modulus := new(big.Int).Set(moduli[0])

	for i := 1; i < len(residues); i++ {
		r2 := residues[i]
		m2 := moduli[i]

		// Reduce r2 modulo m2 so the subtraction below stays small.
		r2mod := new(big.Int).Mod(r2, m2)

		diff := new(big.Int).Sub(r2mod, result)
		diff.Mod(diff, m2)

		inv := new(big.Int).ModInverse(modulus, m2)
		if inv == nil {
			// The pairwise check above guarantees this cannot happen
			// for coprime moduli, but guard defensively.
			return nil, ErrNoSolution
		}

		t := new(big.Int).Mul(diff, inv)
		t.Mod(t, m2)

		step := new(big.Int).Mul(modulus, t)
		result.Add(result, step)
		modulus.Mul(modulus, m2)
		result.Mod(result, modulus)
	}

	return result, nil
}
