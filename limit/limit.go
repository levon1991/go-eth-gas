package limit

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/levon1991/go-eth-gas/limit/hold"

	log "github.com/sirupsen/logrus"
	"github.com/ybbus/jsonrpc/v2"
)

type Limiter interface {
	Gas() int64
	Start()
}

type Conn struct {
	URL      string
	Username string
	Password string
	client   jsonrpc.RPCClient
}

func (c *Conn) Connect() {
	url := fmt.Sprintf("https://%s", c.URL)
	c.client = jsonrpc.NewClientWithOpts(url, &jsonrpc.RPCClientOpts{
		CustomHeaders: map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString(
				[]byte(c.Username+":"+c.Password)),
		},
	})

	if c.client == nil {
		log.Fatalf("Can't connect to wallet")
	}
}

func (c Conn) Start() {
	c.Connect()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		fee, err := hold.GetEstimateLimit(c.client)
		if err != nil {
			return
		}
		log.Info("fee: ", fee)
	}

}

func (c Conn) Gas() int64 {
	return 0
}
