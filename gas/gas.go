package gas

import (
	"errors"
	"io"
	"net/http"
	"time"

	json "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
)

type gasPrice struct {
	SafeLow int64 `json:"safeLow"`
}

var (
	gasPriceUrl        = "https://ethgasstation.info/api/ethgasAPI.json"
	gasPriceReserveUrl = "https://data-api.defipulse.com/api/v1/egs/api/ethgasAPI.json"
)

type GasService interface {
	GetSafeLow() (price int64, err error)
}

type Gas struct{}

func (g Gas) GetSafeLow() (price int64, err error) {
	return SafeLow()
}

func getSafeLow() int64 {
	res, err := http.Get(gasPriceUrl)
	if err != nil {
		res, err = http.Get(gasPriceReserveUrl)
		if err != nil {
			return -1
		}
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	gas := gasPrice{}
	if err = json.NewDecoder(res.Body).Decode(&gas); err != nil {
		return -1
	}

	return gas.SafeLow
}

func getSafeLowByTicker() (int64, error) {
	ticker := time.NewTicker(5 * time.Second)
	ticker2 := time.NewTicker(11 * time.Second)
	for {
		select {
		case <-ticker.C:
			low := getSafeLow()
			if low == -1 {
				log.Info().Msg("can't get gas second time")
				continue
			}
			return low, nil
		case <-ticker2.C:
			return 0, errors.New("can't get optimal gas gas")
		}
	}
}

var GWeiMultiplier int64 = 1000000000

func SafeLow() (price int64, err error) {
	safeLow := getSafeLow()
	if safeLow == -1 {
		log.Info().Msg("can't get gas first time")
		if safeLow, err = getSafeLowByTicker(); err != nil {
			log.Info().Msg("can't get gas last time")
			return 0, err
		}
	}
	log.Info().Msgf("safe low = %d", safeLow)
	gasFeeCheck := safeLow/10 + 10
	price = gasFeeCheck * GWeiMultiplier
	return
}
