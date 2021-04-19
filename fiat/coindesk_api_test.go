package fiat

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

// TestParseCoinDeskData adds a test which checks that we appropriately parse
// the price and timestamp data returned by coindesk's api.
func TestParseCoinDeskData(t *testing.T) {
	var (
		// Create two prices, one which is a float to ensure that we
		// are correctly parsing them.
		price1F         = 10.1
		price2F float64 = 10000

		price1D = decimal.NewFromFloat(price1F)
		price2D = decimal.NewFromFloat(price2F)

		date1 = "2021-04-16"
		date2 = "2021-04-17"
	)

	timestamp1, err := time.Parse(coinDeskTimeFormat, date1)
	require.NoError(t, err)

	timestamp2, err := time.Parse(coinDeskTimeFormat, date2)
	require.NoError(t, err)

	// Create the struct we expect to receive from coindesk and marshal it
	// into bytes.
	resps := coinDeskResponse{
		Data: map[string]float64{
			date1: price1F,
			date2: price2F,
		},
	}

	bytes, err := json.Marshal(resps)
	require.NoError(t, err)

	prices, err := parseCoinDeskData(bytes)
	require.NoError(t, err)

	expectedPrices := []*USDPrice{
		{
			Price:     price1D,
			Timestamp: timestamp1,
		},
		{
			Price:     price2D,
			Timestamp: timestamp2,
		},
	}

	require.Equal(t, expectedPrices, prices)
}
