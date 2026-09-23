package rounding

import (
	"errors"
	"math/big"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

type Mode string

const (
	Truncate Mode = "TRUNCATE"
	HalfUp   Mode = "HALF_UP"
	HalfEven Mode = "HALF_EVEN"
)

func (m Mode) IsValid() bool {
	switch m {
	case Truncate, HalfUp, HalfEven:
		return true
	}
	return false
}

var ErrDivideByZero = errors.New("rounding: cannot divide by zero")

// Divide calculates numerator/denominator with the given decimals, rounding whatever does not fit with the given mode
// We rescale both sides so the whole division is done over the integer of each Amount (an Amount keeps all its digits in one integer plus the number of decimals, so 1.125 is 1125 with 3 decimals)
func Divide(numerator, denominator amount.Amount, decimals int, mode Mode) (*amount.Amount, error) {
	if denominator.IsZero() {
		return nil, ErrDivideByZero
	}

	dividendPower := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(denominator.Decimals()+decimals)), nil)
	dividend := new(big.Int).Mul(numerator.RawValue(), dividendPower)

	divisorPower := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(numerator.Decimals())), nil)
	divisor := new(big.Int).Mul(denominator.RawValue(), divisorPower)

	quotient, remainder := new(big.Int).QuoRem(dividend, divisor, new(big.Int))

	if roundsAwayFromZero(quotient, remainder, divisor, mode) {
		direction := int64(numerator.RawValue().Sign() * denominator.RawValue().Sign())
		quotient.Add(quotient, big.NewInt(direction))
	}

	return amount.New(quotient, decimals)
}

// Round reduces the value to the given decimals, or scales it up if it has fewer
func Round(value amount.Amount, decimals int, mode Mode) (*amount.Amount, error) {
	return Divide(value, *amount.One(), decimals, mode)
}

// roundsAwayFromZero says if the remainder of the division is enough to move the quotient one unit further from zero
func roundsAwayFromZero(quotient, remainder, divisor *big.Int, mode Mode) bool {
	if mode == Truncate || remainder.Sign() == 0 {
		return false
	}

	// We double the remainder and compare it with the divisor to see which side of the half it falls on
	// Absolute values because either side of the division can be negative
	twiceRemainder := new(big.Int).Abs(new(big.Int).Mul(remainder, big.NewInt(2)))
	switch twiceRemainder.Cmp(new(big.Int).Abs(divisor)) {
	case 1:
		return true
	case -1:
		return false
	}

	// HALF_UP always moves away from zero and HALF_EVEN only if the quotient is odd, so that it ends up even
	return mode == HalfUp || isOdd(quotient)
}

func isOdd(value *big.Int) bool {
	return value.Bit(0) == 1
}
