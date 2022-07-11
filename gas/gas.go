package gas

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type GasService interface {
	GetSafeLow() int
}

type Gas struct {
	safeLowGlob int
	mu          sync.Mutex
	Ch          chan struct{}
}

func (g *Gas) GetSafeLow() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.safeLowGlob
}

type Price struct {
	Result struct {
		SafeGasPrice string `json:"SafeGasPrice"`
	} `json:"result"`
}

var gasPriceURL = "https://api.etherscan.io/api?module=gastracker&action=gasoracle&apikey="

func (g *Gas) SafeLowByTicker() {
	res, err := http.Get(gasPriceURL)
	if err != nil {
		return
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	gas := Price{}
	err = json.NewDecoder(res.Body).Decode(&gas)
	if err != nil {
		return
	}

	safe, err := strconv.Atoi(gas.Result.SafeGasPrice)
	if err != nil {
		return
	}

	g.mu.Lock()
	g.safeLowGlob = safe
	g.mu.Unlock()
}

func StartTicker(f func(), delay int) chan struct{} {
	done := make(chan struct{}, 1)
	go func() {
		ticker := time.NewTicker(time.Second * time.Duration(delay))
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				f()
			case <-done:
				return
			}
		}
	}()
	return done
}

func New(delay int) *Gas {
	gas := &Gas{}
	gas.SafeLowByTicker()
	gas.Ch = StartTicker(func() {
		gas.SafeLowByTicker()
	}, delay)
	return gas
}
