package hold

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	json "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	"github.com/ybbus/jsonrpc/v2"
	"golang.org/x/sync/errgroup"
)

type EthTransactions struct {
	GasPrice string `json:"gasPrice"`
	Input    string `json:"input"`
	Hash     string `json:"hash"`
	To       string `json:"to"`
}

var ErrorIncorrectDataFormat = "incorrect data format"
var WeiEthMultiplier int64 = 1000000000
var (
	usdTContractAddress = "0xdac17f958d2ee523a2206206994597c13d831ec7"
	usdCContractAddress = "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"
	erc20               = "0xa9059cbb"
)

func getLastBlockNumber(client jsonrpc.RPCClient) (string, error) {
	result, err := requestCall(client, "eth_blockNumber")
	if err != nil {
		return "", err
	}

	switch num := result.Result.(type) {
	case string:
		return num, nil
	default:
		return "", errors.New(ErrorIncorrectDataFormat)
	}
}

func getTransactionsFromBlock(client jsonrpc.RPCClient, taskChen chan float64, hexNum string) error {
	result, err := requestCall(client, "eth_getBlockByNumber", hexNum, true)
	if err != nil {
		return nil
	}

	gasLimitList := make(map[string]float64)

	switch block := result.Result.(type) {
	case map[string]interface{}:
		sum := 0.0
		count := 0
		for _, v := range block["transactions"].([]interface{}) {
			trans := v.(map[string]interface{})
			transJSON, _ := json.Marshal(trans)
			tr := &EthTransactions{}
			if err := json.Unmarshal(transJSON, &tr); err != nil {
				log.Fatal().Msg(err.Error())
			}
			if tr.Input != "0x" && len(tr.Input) > 10 && tr.Input[:10] == erc20 && (tr.To == usdTContractAddress || tr.To == usdCContractAddress) {
				gasRecipient := getRecipient(client, tr.Hash)
				if gasRecipient == "" {
					continue
				}

				gas, gasPrice, err := gasHexToDecimal(gasRecipient, tr.GasPrice)
				if err != nil {
					continue
				}

				maxFeeWei := gas * gasPrice / 1000000000

				maxFeeEt := float64(maxFeeWei) / float64(WeiEthMultiplier)
				sum += maxFeeEt
				count++
				gasLimitList[tr.Hash] = maxFeeEt
			}
		}
		if count > 0 {
			taskChen <- sum / float64(count)
		}

		return nil
	default:
		return nil
	}
}

func gasHexToDecimal(gasRecipient string, gesPrice string) (int64, int64, error) {
	gas, err := strconv.ParseInt(strings.ReplaceAll(gasRecipient, "0x", ""), 16, 64)
	if err != nil {
		return 0, 0, err
	}
	gasPrice, err := strconv.ParseInt(strings.ReplaceAll(gesPrice, "0x", ""), 16, 64)
	if err != nil {
		return 0, 0, err
	}
	return gas, gasPrice, nil
}

func getRecipient(client jsonrpc.RPCClient, hash string) string {
	transaction, err := requestCall(client, "eth_getTransactionReceipt", hash)
	if err != nil {
		return ""
	}

	var gasRecipient string
	if recipient, ok := transaction.Result.(map[string]interface{}); ok {
		if gasUsed, ok := recipient["gasUsed"]; ok {
			if gasRecipient, ok = gasUsed.(string); ok {
				return gasRecipient
			}
		}
	}
	return gasRecipient
}

func GetEstimateLimit(client jsonrpc.RPCClient) (float64, error) {
	blockNumHex, err := getLastBlockNumber(client)
	if err != nil {
		return 0, err
	}

	blockNum, err := strconv.ParseInt(strings.ReplaceAll(blockNumHex, "0x", ""), 16, 64)
	if err != nil {
		return 0, err
	}
	fmt.Println("block Num = ", blockNum)

	taskChan := make(chan float64)
	wg := &errgroup.Group{}
	for i := blockNum - 10; i < blockNum; i++ {
		i := i
		wg.Go(func() error {
			return getTransactionsFromBlock(client, taskChan, fmt.Sprintf("0x%x", i))
		})
	}

	go func() {
		if err = wg.Wait(); err != nil {
			log.Error().Msg(err.Error())
		}
		close(taskChan)
	}()

	count := 0
	globalSum := 0.0
	for sum := range taskChan {
		count++
		globalSum += sum
	}

	return globalSum / float64(count), nil
}
func requestCall(client jsonrpc.RPCClient, methode string, params ...interface{}) (*jsonrpc.RPCResponse, error) {
	response, err := client.Call(methode, params)
	if err != nil {
		return &jsonrpc.RPCResponse{}, err
	}

	if response.Error != nil {
		return &jsonrpc.RPCResponse{}, response.Error
	}

	if response.Result == nil {
		return &jsonrpc.RPCResponse{}, errors.New("empty response")
	}
	return response, nil
}
