package rounding

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

func amt(t *testing.T, value string) amount.Amount {
	t.Helper()
	a, err := amount.NewFromString(value)
	require.NoError(t, err)
	return *a
}

func TestRound(t *testing.T) {
	cases := []struct {
		mode     Mode
		value    string
		decimals int
		expected string
	}{
		// TRUNCATE drops whatever does not fit, even an exact half
		{Truncate, "1.129", 2, "1.12"},
		{Truncate, "1.125", 2, "1.12"},
		{Truncate, "0.009", 2, "0.00"},

		// HALF_UP moves away from zero from the half onwards
		{HalfUp, "1.124", 2, "1.12"},
		{HalfUp, "1.125", 2, "1.13"},
		{HalfUp, "1.126", 2, "1.13"},
		{HalfUp, "0.005", 2, "0.01"},
		{HalfUp, "1.999", 2, "2.00"},

		// HALF_EVEN only moves on an exact half when the digit it leaves behind is odd
		{HalfEven, "1.125", 2, "1.12"},
		{HalfEven, "1.135", 2, "1.14"},
		{HalfEven, "1.126", 2, "1.13"},
		{HalfEven, "1.995", 2, "2.00"},

		// Whole units, where the tie sits on the .5
		{Truncate, "1.5", 0, "1"},
		{HalfUp, "1.5", 0, "2"},
		{HalfEven, "1.5", 0, "2"},
		{HalfEven, "2.5", 0, "2"},

		// Negative amounts move away from zero, or towards it when truncating
		{Truncate, "-1.129", 2, "-1.12"},
		{HalfUp, "-1.125", 2, "-1.13"},
		{HalfEven, "-1.125", 2, "-1.12"},
		{HalfEven, "-1.135", 2, "-1.14"},

		// A value already at the precision, or with fewer decimals, is left alone or scaled up
		{Truncate, "1.12", 2, "1.12"},
		{Truncate, "1.1", 2, "1.10"},
		{Truncate, "7", 2, "7.00"},
		{Truncate, "0", 2, "0.00"},

		// DLT amounts keep their size and precision
		{Truncate, "0.3333333333333333333333333333", 18, "0.333333333333333333"},
		{HalfUp, "333333333333333333333333.335", 2, "333333333333333333333333.34"},
	}

	for _, c := range cases {
		result, err := Round(amt(t, c.value), c.decimals, c.mode)

		require.NoError(t, err)
		require.Equal(t, c.expected, result.String(), "%s of %s to %d decimals", c.mode, c.value, c.decimals)
	}
}

func TestDivide(t *testing.T) {
	cases := []struct {
		mode        Mode
		numerator   string
		denominator string
		decimals    int
		expected    string
	}{
		{Truncate, "100", "8", 2, "12.50"},
		{Truncate, "100", "7", 4, "14.2857"},
		{HalfUp, "100", "7", 4, "14.2857"},
		{Truncate, "100", "3", 0, "33"},
		{HalfUp, "100", "6", 0, "17"},

		// 5/2 lands on an exact half, so the mode decides
		{Truncate, "5", "2", 0, "2"},
		{HalfUp, "5", "2", 0, "3"},
		{HalfEven, "5", "2", 0, "2"},

		{Truncate, "1", "3", 18, "0.333333333333333333"},

		// Decimals on both sides of the division
		{Truncate, "10.25", "2.5", 2, "4.10"},
		{Truncate, "1.5", "0.5", 2, "3.00"},
		{HalfUp, "10.5", "0.7", 3, "15.000"},

		// A negative on either side flips the result, and the rounding still moves away from zero
		{Truncate, "100", "-8", 2, "-12.50"},
		{Truncate, "-100", "-8", 2, "12.50"},
		{HalfUp, "-5", "2", 0, "-3"},
		{HalfUp, "5", "-2", 0, "-3"},
		{HalfEven, "-5", "2", 0, "-2"},
	}

	for _, c := range cases {
		result, err := Divide(amt(t, c.numerator), amt(t, c.denominator), c.decimals, c.mode)

		require.NoError(t, err)
		require.Equal(t, c.expected, result.String(), "%s of %s/%s to %d decimals", c.mode, c.numerator, c.denominator, c.decimals)
	}
}

func TestDivideByZero(t *testing.T) {
	for _, denominator := range []string{"0", "0.00", "-0"} {
		_, err := Divide(amt(t, "100"), amt(t, denominator), 2, Truncate)

		require.ErrorIs(t, err, ErrDivideByZero, "denominator %s", denominator)
	}
}
