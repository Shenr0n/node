package config

import (
	"context"

	"google.golang.org/grpc"

	"github.com/btcsuite/btcd/rpcclient"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/zeta-chain/zetacore/contrib/localnet/orchestrator/smoketest/config"
	crosschaintypes "github.com/zeta-chain/zetacore/x/crosschain/types"
	fungibletypes "github.com/zeta-chain/zetacore/x/fungible/types"
	observertypes "github.com/zeta-chain/zetacore/x/observer/types"
)

// getClientsFromConfig get clients from config
func getClientsFromConfig(ctx context.Context, conf config.Config, evmPrivKey string) (
	*rpcclient.Client,
	*ethclient.Client,
	*bind.TransactOpts,
	crosschaintypes.QueryClient,
	fungibletypes.QueryClient,
	authtypes.QueryClient,
	banktypes.QueryClient,
	observertypes.QueryClient,
	*ethclient.Client,
	*bind.TransactOpts,
	error,
) {
	btcRPCClient, err := getBtcClient(conf.RPCs.Bitcoin)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	goerliClient, goerliAuth, err := getEVMClient(ctx, conf.RPCs.EVM, evmPrivKey)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	cctxClient, fungibleClient, authClient, bankClient, observerClient, err := getZetaClients(conf.RPCs.ZetaCoreGRPC)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	zevmClient, zevmAuth, err := getEVMClient(ctx, conf.RPCs.Zevm, evmPrivKey)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	return btcRPCClient,
		goerliClient,
		goerliAuth,
		cctxClient,
		fungibleClient,
		authClient,
		bankClient,
		observerClient,
		zevmClient,
		zevmAuth,
		nil
}

// getBtcClient get btc client
func getBtcClient(rpc string) (*rpcclient.Client, error) {
	connCfg := &rpcclient.ConnConfig{
		Host:         rpc,
		User:         "smoketest",
		Pass:         "123",
		HTTPPostMode: true,
		DisableTLS:   true,
		Params:       "testnet3",
	}
	return rpcclient.New(connCfg, nil)
}

// getEVMClient get goerli client
func getEVMClient(ctx context.Context, rpc, privKey string) (*ethclient.Client, *bind.TransactOpts, error) {
	evmClient, err := ethclient.Dial(rpc)
	if err != nil {
		return nil, nil, err
	}

	chainid, err := evmClient.ChainID(ctx)
	if err != nil {
		return nil, nil, err
	}
	deployerPrivkey, err := crypto.HexToECDSA(privKey)
	if err != nil {
		return nil, nil, err
	}
	evmAuth, err := bind.NewKeyedTransactorWithChainID(deployerPrivkey, chainid)
	if err != nil {
		return nil, nil, err
	}

	return evmClient, evmAuth, nil
}

// getZetaClients get zeta clients
func getZetaClients(rpc string) (
	crosschaintypes.QueryClient,
	fungibletypes.QueryClient,
	authtypes.QueryClient,
	banktypes.QueryClient,
	observertypes.QueryClient,
	error,
) {
	grpcConn, err := grpc.Dial(rpc, grpc.WithInsecure())
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	cctxClient := crosschaintypes.NewQueryClient(grpcConn)
	fungibleClient := fungibletypes.NewQueryClient(grpcConn)
	authClient := authtypes.NewQueryClient(grpcConn)
	bankClient := banktypes.NewQueryClient(grpcConn)
	observerClient := observertypes.NewQueryClient(grpcConn)

	return cctxClient, fungibleClient, authClient, bankClient, observerClient, nil
}
