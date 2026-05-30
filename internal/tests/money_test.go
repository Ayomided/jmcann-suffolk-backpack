package tests

import (
	"testing"

	"adediiji.uk/jmcann-suffolk-backpack-task/internal/model"
)

func TestMoneyDisplayUSD(t *testing.T) {
	m := model.NewMoney("usd", 10000)
	if m.String() != "$100.00" {
		t.Errorf("expected $100.00 got %s", m.String())
	}
}

func TestMoneyDisplayGBP(t *testing.T) {
	m := model.NewMoney("gbp", 1050)
	if m.String() != "£10.50" {
		t.Errorf("expected £10.50 got %s", m.String())
	}
}

func TestMoneyDisplayEuro(t *testing.T) {
	m := model.NewMoney("euro", 2000)
	if m.String() != "20.00€" {
		t.Errorf("expected 20.00€ got %s", m.String())
	}
}

func TestMoneyAddSameCurrency(t *testing.T) {
	a := model.NewMoney("gbp", 1000)
	b := model.NewMoney("gbp", 500)
	result, err := a.Add(b)
	if err != nil {
		t.Fatal(err)
	}
	if result.Amount() != 1500 {
		t.Errorf("expected 1500 got %d", result.Amount())
	}
}

func TestMoneyAddDifferentCurrencyReturnsError(t *testing.T) {
	a := model.NewMoney("gbp", 1000)
	b := model.NewMoney("usd", 1000)
	_, err := a.Add(b)
	if err == nil {
		t.Error("expected error when adding different currencies")
	}
}

func TestMoneyMultiplyForRateCalculation(t *testing.T) {
	rate := model.NewMoney("gbp", 1000)
	total := rate.Multiply(2.5)
	if total.Amount() != 2500 {
		t.Errorf("expected 2500 got %d", total.Amount())
	}
}

func TestMoneyMultiplyForRateQCalculation(t *testing.T) {
	rate := model.NewMoney("gbp", 12000)
	total := rate.Multiply(10.0)
	if total.Amount() != 120000 {
		t.Errorf(" expected 120000 got %d %s", total.Amount(), total.String())
	}
}
