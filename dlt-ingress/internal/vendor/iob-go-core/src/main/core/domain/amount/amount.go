package amount

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

var ErrValueCannotBeNil = errors.New("value cannot be nil")

func New(value *big.Int, decimals int) (*Amount, error) {
	if value == nil {
		return nil, ErrValueCannotBeNil
	}
	return &Amount{value: value, decimals: decimals}, nil
}

var ErrInvalidAmountString = errors.New("invalid amount string")

var validAmountStr = regexp.MustCompile(`^([+|-])?(\d*)\.?(\d*)$`)

func NewFromString(str string) (*Amount, error) {

	if str == "" || !validAmountStr.MatchString(str) {
		return nil, ErrInvalidAmountString
	}

	var value, _ = new(big.Int).SetString(validAmountStr.ReplaceAllString(str, "$1$2$3"), 10)
	var decimals = len(validAmountStr.ReplaceAllString(str, "$3"))

	return &Amount{value: value, decimals: decimals}, nil
}

type Amount struct {
	value    *big.Int
	decimals int
}

func (a Amount) Decimals() int {
	return a.decimals
}

/* WARN: This method returns the raw data of Amount. F.e 1.25 will return 125. not 1 */
func (a Amount) RawValue() *big.Int {
	return a.value
}

func (a Amount) String() string {
	if a.value == nil {
		return ""
	}

	if a.decimals == 0 {
		return a.value.String()
	}

	var sign string
	if a.value.Sign() == -1 {
		sign = "-"
	}

	absValueStr := new(big.Int).Abs(a.value).String()
	absValueStrLength := len(absValueStr)
	if absValueStrLength <= a.decimals {
		format := fmt.Sprintf("%%s0.%%%ds", a.decimals)
		return strings.ReplaceAll(fmt.Sprintf(format, sign, absValueStr), " ", "0")
	}
	index := absValueStrLength - a.decimals
	return fmt.Sprintf("%s%s.%s", sign, absValueStr[:index], absValueStr[index:])
}

func (a Amount) MarshalText() ([]byte, error) {
	amountStr := a.String()
	return []byte(amountStr), nil
}

func (a *Amount) UnmarshalText(data []byte) error {
	newAmount, err := NewFromString(string(data))
	if err != nil {
		return err
	}
	*a = *newAmount
	return nil
}

func (a Amount) MarshalJSON() ([]byte, error) {
	amountStr := a.String()
	data, err := json.Marshal(amountStr)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (a *Amount) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		newAmount, err := NewFromString(str)
		if err != nil {
			return err
		}
		*a = *newAmount
		return nil
	}
	return fmt.Errorf("amount: unsupported JSON value %s", string(data))
}

func (a Amount) Add(amounts ...Amount) Amount {
	var result = copyAmount(a)

	for _, amount := range amounts {
		result = add(result, amount)
	}

	return result
}

func (a Amount) Sub(amounts ...Amount) Amount {
	var result = copyAmount(a)

	for _, amount := range amounts {
		result = sub(result, amount)
	}

	return result
}

func (a Amount) Mul(amounts ...Amount) Amount {
	var result = copyAmount(a)

	var decimals = result.decimals
	for _, amount := range amounts {
		if amount.decimals > decimals {
			decimals = amount.decimals
		}
		result = mul(result, amount)
	}

	result = setDecimals(result, decimals)

	return result
}

func (a Amount) Rem(amounts ...Amount) Amount {
	var result = copyAmount(a)

	for _, amount := range amounts {
		result = rem(result, amount)
	}

	return result
}

func (a Amount) MulWithDecimals(amount Amount, decimals int) Amount {
	return setDecimals(mul(copyAmount(a), amount), decimals)
}

func (a Amount) Div(amounts ...Amount) Amount {
	return a.Quo(amounts...)
}

func (a Amount) DivWithDecimals(amount Amount, decimals int) Amount {
	return a.QuoWithDecimals(amount, decimals)
}

func (a Amount) Quo(amounts ...Amount) Amount {
	var result = copyAmount(a)

	for _, amount := range amounts {
		result = quo(result, amount)
	}

	return result
}

func (a Amount) QuoWithDecimals(amount Amount, decimals int) Amount {
	return quoWithDecimals(copyAmount(a), amount, decimals)
}

func (a Amount) WithDecimals(decimals int) (*Amount, error) {
	if decimals < a.decimals {
		errMsg := fmt.Sprintf("error setting decimals to amount: the number of decimals in [%s] is higher than [%d]", a.String(), decimals)
		return nil, errors.New(errMsg)
	}

	var powerBase = big.NewInt(10)
	var powerExponent = new(big.Int).Sub(big.NewInt(int64(decimals)), big.NewInt(int64(a.decimals)))
	var power = new(big.Int).Exp(powerBase, powerExponent, nil)
	return &Amount{value: new(big.Int).Mul(a.value, power), decimals: decimals}, nil
}

func (a *Amount) FormatAmount(supportedDecimals int) error {
	if a.decimals > supportedDecimals {
		errMsg := fmt.Sprintf("error setting formating amount: the number of decimals of the provided amount [%s] is higher than supported one [%d]", a.String(), supportedDecimals)
		return errors.New(errMsg)
	}
	if a.decimals == supportedDecimals {
		return nil
	}
	if a.decimals < supportedDecimals {
		formattedAmount, err := a.WithDecimals(supportedDecimals)
		if err != nil {
			return err
		}
		a.decimals = formattedAmount.decimals
		a.value = formattedAmount.value
		return nil
	}
	return nil
}

func add(a, b Amount) Amount {
	var x, y, decimals = matchDecimals(a, b)
	var sum = new(big.Int).Add(x.value, y.value)
	return Amount{value: sum, decimals: decimals}
}

func sub(a, b Amount) Amount {
	var x, y, decimals = matchDecimals(a, b)
	var diff = new(big.Int).Sub(x.value, y.value)
	return Amount{value: diff, decimals: decimals}
}

func mul(a, b Amount) Amount {
	var product = new(big.Int).Mul(a.value, b.value)
	var decimals = a.decimals + b.decimals
	return Amount{value: product, decimals: decimals}
}

func quo(a, b Amount) Amount {
	var decimals = getHigherDecimals(a, b)
	return quoWithDecimals(a, b, decimals)
}

func rem(a, b Amount) Amount {
	var x, y, decimals = matchDecimals(a, b)
	var rem = new(big.Int).Rem(x.value, y.value)
	return Amount{value: rem, decimals: decimals}
}

func quoWithDecimals(a, b Amount, decimals int) Amount {
	var dividendPowerBase = big.NewInt(10)
	var dividendPowerExponent = new(big.Int).Add(big.NewInt(int64(b.decimals)), big.NewInt(int64(decimals)))
	var dividendPower = new(big.Int).Exp(dividendPowerBase, dividendPowerExponent, nil)
	var dividend = new(big.Int).Mul(a.value, dividendPower)

	var divisorPowerBase = big.NewInt(10)
	var divisorPowerExponent = big.NewInt(int64(a.decimals))
	var divisorPower = new(big.Int).Exp(divisorPowerBase, divisorPowerExponent, nil)
	var divisor = new(big.Int).Mul(b.value, divisorPower)

	// quotient = a.value * 10 ^ (b.decimals + decimals) / (b.value * 10 ^ a.decimals)
	var quotient = new(big.Int).Quo(dividend, divisor)

	return Amount{value: quotient, decimals: decimals}
}

func setDecimals(a Amount, decimals int) Amount {
	switch {
	case a.decimals > decimals:
		var decimalsToRemove = new(big.Int).Sub(big.NewInt(int64(a.decimals)), big.NewInt(int64(decimals)))
		var currentAmountStr = a.String()
		var index = int64(len(currentAmountStr)) - decimalsToRemove.Int64()
		var newAmountStr = currentAmountStr[:index]
		amount, _ := NewFromString(newAmountStr)
		return *amount
	case a.decimals < decimals:
		var powerBase = big.NewInt(10)
		var powerExponent = new(big.Int).Sub(big.NewInt(int64(decimals)), big.NewInt(int64(a.decimals)))
		var power = new(big.Int).Exp(powerBase, powerExponent, nil)
		return Amount{value: new(big.Int).Mul(a.value, power), decimals: decimals}
	default:
		return a
	}
}

func getHigherDecimals(a, b Amount) int {
	switch {
	case a.decimals > b.decimals:
		return a.decimals
	case a.decimals < b.decimals:
		return b.decimals
	default:
		return a.decimals
	}
}

func matchDecimals(a, b Amount) (Amount, Amount, int) {
	switch {
	case a.decimals > b.decimals:
		var powerBase = big.NewInt(10)
		var powerExponent = new(big.Int).Sub(big.NewInt(int64(a.decimals)), big.NewInt(int64(b.decimals)))
		var power = new(big.Int).Exp(powerBase, powerExponent, nil)
		return a, Amount{value: new(big.Int).Mul(b.value, power), decimals: a.decimals}, a.decimals
	case a.decimals < b.decimals:
		var powerBase = big.NewInt(10)
		var powerExponent = new(big.Int).Sub(big.NewInt(int64(b.decimals)), big.NewInt(int64(a.decimals)))
		var power = new(big.Int).Exp(powerBase, powerExponent, nil)
		return Amount{value: new(big.Int).Mul(a.value, power), decimals: b.decimals}, b, b.decimals
	default:
		return a, b, a.decimals
	}
}

func copyAmount(a Amount) Amount {
	return Amount{
		value:    new(big.Int).Set(a.value),
		decimals: a.decimals,
	}
}

func (a Amount) IsZero() bool {
	dv, _ := a.Value()
	return isZeroDriverValue(dv)
}

func (a Amount) Float64() (float64, error) {
	dv, err := a.Value()
	if err != nil {
		return 0, err
	}
	return float64FromDriverValue(dv)
}

func isZeroDriverValue(dv driver.Value) bool {
	switch v := dv.(type) {
	case nil:
		return true
	case int64:
		return v == 0
	case float64:
		return v == 0
	case string:
		return isZeroText(v)
	case []byte:
		return isZeroText(string(v))
	default:
		return false
	}
}

func isZeroText(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return true
	}
	s = strings.TrimPrefix(s, "-")

	s = strings.TrimLeft(s, "0")
	s = strings.TrimPrefix(s, ".")
	s = strings.TrimRight(s, "0")
	return s == ""
}

func float64FromDriverValue(dv driver.Value) (float64, error) {
	switch v := dv.(type) {
	case nil:
		return 0, fmt.Errorf("amount: null value")
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		return float64FromString(v)
	case []byte:
		return float64FromString(string(v))
	default:
		return 0, fmt.Errorf("amount: unsupported driver.Value type %T", v)
	}
}

func float64FromString(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("amount: empty string")
	}
	var r big.Rat
	if _, ok := r.SetString(s); !ok {
		return 0, fmt.Errorf("amount: invalid decimal string %q", s)
	}
	f, _ := r.Float64()
	return f, nil
}

func (a Amount) Cmp(b Amount) int {
	x, y, _ := matchDecimals(a, b)
	return x.value.Cmp(y.value)
}

func (a Amount) LessThan(b Amount) bool    { return a.Cmp(b) < 0 }
func (a Amount) GreaterThan(b Amount) bool { return a.Cmp(b) > 0 }
func (a Amount) Equal(b Amount) bool       { return a.Cmp(b) == 0 }

func Zero() *Amount {
	return &Amount{
		value:    big.NewInt(0),
		decimals: 0,
	}
}

func One() *Amount {
	return &Amount{
		value:    big.NewInt(1),
		decimals: 0,
	}
}

func Ten() *Amount {
	return &Amount{
		value:    big.NewInt(10),
		decimals: 0,
	}
}

func Hundred() *Amount {
	return &Amount{
		value:    big.NewInt(100),
		decimals: 0,
	}
}

func TwoHundred() *Amount {
	return &Amount{
		value:    big.NewInt(200),
		decimals: 0,
	}
}
