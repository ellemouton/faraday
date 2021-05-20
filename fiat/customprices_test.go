package fiat

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/stretchr/testify/require"
)

// TestCustomRawPriceData tests that price data in the CSV format is parsed
// correctly.
func TestCustomRawPriceData(t *testing.T) {
	var (
		// Create two prices, one which is a float to ensure that we
		// are correctly parsing them.
		price1 = decimal.NewFromFloat(10.1)
		price2 = decimal.NewFromInt(110000)

		// Create two timestamps, each representing our time in
		// milliseconds.
		time1 = time.Unix(10000, 0)
		time2 = time.Unix(2000, 0)
	)

	getDataFunc := func(string) ([][]string, error) {
		return [][]string{
			{strconv.FormatInt(time1.Unix(), 10), price1.String()},
			{strconv.FormatInt(time2.Unix(), 10), price2.String()},
		}, nil
	}

	c := &customPrices{
		currency: "USD",
		getData:  getDataFunc,
	}

	prices, err := c.rawPriceData(context.Background(), time.Time{}, time.Time{})
	require.NoError(t, err)

	expectedPrices := []*Price{
		{
			Timestamp: time1,
			Price:     price1,
			Currency:  "USD",
		},
		{
			Timestamp: time2,
			Price:     price2,
			Currency:  "USD",
		},
	}

	for i, p := range expectedPrices {
		require.True(t, p.Price.Equal(prices[i].Price))
		require.True(t, p.Timestamp.Equal(prices[i].Timestamp))
		require.Equal(t, p.Currency, prices[i].Currency)
	}
}
