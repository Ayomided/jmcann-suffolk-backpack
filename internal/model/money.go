package model

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type CostUnit string

const (
	CostUnitPerHour CostUnit = "per_hour"
	CostUnitPerUnit CostUnit = "per_unit"
)

type Currency string

const (
	CurrencyGBP  Currency = "GBP"
	CurrencyUSD  Currency = "USD"
	CurrencyEuro Currency = "EURO"
)

func CurrencyFrom(s string) Currency {
	switch s {
	case "USD", "Usd", "usd":
		return CurrencyUSD
	case "EURO", "Euro", "euro":
		return CurrencyEuro
	default:
		return CurrencyGBP
	}
}

func (c Currency) Symbol() string {
	switch c {
	case CurrencyUSD:
		return "$"
	case CurrencyEuro:
		return "€"
	default:
		return "£"
	}
}

func CurrencySymbol(symbol string) Currency {
	switch symbol {
	case "$":
		return CurrencyUSD
	case "€":
		return CurrencyEuro
	default:
		return CurrencyGBP
	}
}

type Money struct {
	amount    uint64
	precision uint32
	currency  Currency
}

func NewMoney(currency string, amount uint64) Money {
	return Money{
		amount:    amount,
		precision: 2,
		currency:  CurrencyFrom(currency),
	}
}

func (m Money) Amount() uint64 {
	return m.amount
}

func (m Money) Currency() Currency {
	return m.currency
}

func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, errors.New("cannot add money of different currencies")
	}
	return Money{
		amount:    m.amount + other.amount,
		precision: m.precision,
		currency:  m.currency,
	}, nil
}

func (m Money) Multiply(factor float64) Money {
	return Money{
		amount:    uint64(math.Round(float64(m.amount) * factor)),
		precision: m.precision,
		currency:  m.currency,
	}
}

func (m Money) String() string {
	decimalValue := float64(m.amount) / math.Pow10(int(m.precision))
	switch m.currency {
	case CurrencyEuro:
		return fmt.Sprintf("%.2f%s", decimalValue, m.currency.Symbol())
	default:
		return fmt.Sprintf("%s%.2f", m.currency.Symbol(), decimalValue)
	}
}

func (m Money) Serialize() string {
	return fmt.Sprintf("%s%d", m.currency.Symbol(), m.amount)
}

func MoneyFromString(value string) (*Money, error) {
	if value == "" {
		return nil, errors.New("empty money string")
	}

	s := strings.Split(value, "")
	if len(s) < 2 {
		return nil, errors.New("invalid money string")
	}

	symbol := s[0]

	switch symbol {
	case "$", "£", "€":
	default:
		return nil, fmt.Errorf("invalid currency symbol: %s", symbol)
	}

	valueStr := strings.Join(s[1:], "")
	moneyAmount, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse money value: %w", err)
	}

	return &Money{
		amount:    uint64(moneyAmount),
		precision: 2,
		currency:  CurrencySymbol(symbol),
	}, nil
}
