package fiat

import (
	"context"
	"encoding/csv"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

// customPrices implements the fiatBackend interface.
type customPrices struct {
	currency string
	csvPath  string

	// getData is the function to be used to fetch the unparsed price data.
	// It is set within this struct so that it can be mocked for testing.
	getData func(string) ([][]string, error)
}

func newCustomPricesAPI(csvPath string, currency string) *customPrices {
	return &customPrices{
		csvPath:  csvPath,
		currency: currency,
		getData:  readDataFromFile,
	}
}

func readDataFromFile(path string) ([][]string, error) {
	csvFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer csvFile.Close()

	return csv.NewReader(csvFile).ReadAll()
}

// rawPriceData fetches the CSV encoded price data and each entry into a Price
// struct.
func (c *customPrices) rawPriceData(_ context.Context, _,
	_ time.Time) ([]*Price, error) {

	csvLines, err := c.getData(c.csvPath)
	if err != nil {
		return nil, err
	}

	prices := make([]*Price, len(csvLines))

	for i, line := range csvLines {
		if len(line) != 2 {
			return nil, errors.New("incorrect csv format")
		}

		timestamp, err := strconv.ParseInt(line[0], 10, 64)
		if err != nil {
			return nil, err
		}

		price, err := strconv.ParseFloat(line[1], 64)
		if err != nil {
			return nil, err
		}

		prices[i] = &Price{
			Timestamp: time.Unix(timestamp, 0),
			Price:     decimal.NewFromFloat(price),
			Currency:  c.currency,
		}
	}

	return prices, nil
}
