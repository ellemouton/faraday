package fiat

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"time"

	"github.com/lightninglabs/faraday/utils"
	"github.com/shopspring/decimal"
)

const (
	// coinDeskHistoryAPI is the endpoint we hit for historical price data.
	coinDeskHistoryAPI = "https://api.coindesk.com/v1/bpi/historical/close.json"

	// coinDeskTimeFormat is the date format used by coindesk.
	coinDeskTimeFormat = "2006-01-02"
)

// coinDeskAPI implements the PriceAPIBackend interface.
type coinDeskAPI struct{}

type coinDeskResponse struct {
	Data map[string]float64 `json:"bpi"`
}

// queryCoinDesk constructs and sends a request to coindesk to query historical
// price information.
func queryCoinDesk(start, end time.Time) ([]byte, error) {
	queryURL := fmt.Sprintf("%v?start=%v&end=%v",
		coinDeskHistoryAPI, start.Format(coinDeskTimeFormat),
		end.Format(coinDeskTimeFormat))

	log.Debugf("coindesk url: %v", queryURL)

	response, err := http.Get(queryURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)
}

// parseCoinDeskData parses http response data from coindesk into USDPrice
// structs.
func parseCoinDeskData(data []byte) ([]*USDPrice, error) {
	var priceEntries coinDeskResponse
	if err := json.Unmarshal(data, &priceEntries); err != nil {
		return nil, err
	}

	var usdRecords []*USDPrice

	for date, price := range priceEntries.Data {
		timestamp, err := time.Parse(coinDeskTimeFormat, date)
		if err != nil {
			return nil, err
		}

		usdRecords = append(usdRecords, &USDPrice{
			Timestamp: timestamp,
			Price:     decimal.NewFromFloat(price),
		})
	}

	return usdRecords, nil
}

// GetPrices retrieves price information from coindesks's api for the given
// time range.
func (c *coinDeskAPI) GetPrices(ctx context.Context, start,
	end time.Time) ([]*USDPrice, error) {

	// First, check that we have a valid start and end time, and that the
	// range specified is not in the future.
	if err := utils.ValidateTimeRange(
		start, end, utils.DisallowFutureRange,
	); err != nil {
		return nil, err
	}

	query := func() ([]byte, error) {
		return queryCoinDesk(start, end)
	}

	// CoinDesk uses a granularity of 1 day and does not include the current
	// day's price information. So subtract 1 period from the start date so
	// that at least one day's price data is always included.
	start = start.Add(time.Hour * -24)

	// Query the api for this page of data. We allow retries at this
	// stage in case the api experiences a temporary limit.
	records, err := retryQuery(ctx, query, parseCoinDeskData)
	if err != nil {
		return nil, err
	}

	// Sort by ascending timestamp once we have all of our records. We
	// expect these records to already be sorted, but we do not trust our
	// external source to do so (just in case).
	sort.SliceStable(records, func(i, j int) bool {
		return records[i].Timestamp.Before(
			records[j].Timestamp,
		)
	})

	return records, nil
}
