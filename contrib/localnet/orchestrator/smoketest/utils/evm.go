package utils

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	DefaultReceiptTimeout = 30 * time.Second
)

func CheckNonce(
	ctx context.Context,
	client *ethclient.Client,
	addr ethcommon.Address,
	expectedNonce uint64,
) error {
	nonce, err := client.PendingNonceAt(ctx, addr)
	if err != nil {
		return err
	}
	if nonce != expectedNonce {
		return fmt.Errorf("want nonce %d; got %d", expectedNonce, nonce)
	}
	return nil
}

// MustWaitForTxReceipt waits until a broadcasted tx to be mined and return its receipt
// timeout and panic after 30s.
func MustWaitForTxReceipt(
	ctx context.Context,
	client *ethclient.Client,
	tx *ethtypes.Transaction,
	logger infoLogger,
	receiptTimeout time.Duration,
) *ethtypes.Receipt {
	timeout := DefaultReceiptTimeout
	if receiptTimeout != 0 {
		timeout = receiptTimeout
	}

	start := time.Now()
	for {
		if time.Since(start) > timeout {
			panic("waiting tx receipt timeout")
		}
		receipt, err := client.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			if !errors.Is(err, ethereum.NotFound) {
				logger.Info("fetching tx receipt error: ", err.Error())
			}
			time.Sleep(1 * time.Second)
			continue
		}
		if receipt != nil {
			return receipt
		}
	}
}

// TraceTx traces the tx and returns the trace result
func TraceTx(ctx context.Context, tx *ethtypes.Transaction, rpcURL string) (string, error) {
	rpcClient, err := rpc.Dial(rpcURL)
	if err != nil {
		return "", err
	}

	var result interface{}
	txHash := tx.Hash().Hex()
	err = rpcClient.CallContext(
		ctx,
		&result,
		"debug_traceTransaction",
		txHash,
		map[string]interface{}{
			"disableMemory":  true,
			"disableStack":   false,
			"disableStorage": false,
			"fullStorage":    false,
		})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Trace result: %+v\n", result), nil
}
