# modular_inverse

Computes modular multiplicative inverses and Chinese remainder theorem solutions over big integers.

```go
package main

import (
	"fmt"
	"math/big"

	"modular_inverse"
)

func main() {
	inv, err := modular_inverse.ModInverse(big.NewInt(3), big.NewInt(11))
	if err != nil {
		panic(err)
	}
	fmt.Println(inv) // 4

	x, err := modular_inverse.CRT(
		[]*big.Int{big.NewInt(2), big.NewInt(3), big.NewInt(2)},
		[]*big.Int{big.NewInt(3), big.NewInt(5), big.NewInt(7)},
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(x) // 23
}
```

## Why this exists

The standard library gives you `big.Int.ModInverse`, but it returns `nil` on failure and makes you remember that `nil` means "no inverse". It has no Chinese remainder theorem helper at all. This package wraps the inverse in a real error and adds a CRT solver that validates its inputs.

The trade-off: `CRT` requires pairwise-coprime moduli and returns a single value in `[0, M)`. That is the common case and keeps the signature simple. If you need non-coprime moduli, you have to reduce the system yourself before calling `CRT`; the package will tell you with `ErrNotCoprime` rather than silently returning a wrong answer.

## The awkward edge

`ModInverse` rejects a modulus of 1. Every integer is congruent to 0 modulo 1, so there is no meaningful inverse. If you pass 1 you get an error, not a zero.

`CRT` also rejects moduli of 1, for the same reason: the product of the moduli would be unchanged by a modulus of 1, and the "unique solution modulo the product" would not be well defined.

## Exported names

- `ModInverse(a, m *big.Int) (*big.Int, error)`
- `CRT(residues, moduli []*big.Int) (*big.Int, error)`
- `ErrNoInverse`
- `ErrNoSolution`
- `ErrNotCoprime`
