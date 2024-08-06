package main

import (
	"context"
	"cosmossdk.io/math"
	"encoding/json"
	"fmt"
	exchangeclient "github.com/InjectiveLabs/sdk-go/client/exchange"
	"os"

	"github.com/InjectiveLabs/sdk-go/client"

	"github.com/InjectiveLabs/sdk-go/client/common"

	exchangetypes "github.com/InjectiveLabs/sdk-go/chain/exchange/types"
	chainclient "github.com/InjectiveLabs/sdk-go/client/chain"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
)

func main() {
	network := common.LoadNetwork("testnet", "lb")
	tmClient, err := rpchttp.New(network.TmEndpoint, "/websocket")
	if err != nil {
		panic(err)
	}

	senderAddress, cosmosKeyring, err := chainclient.InitCosmosKeyring(
		os.Getenv("HOME")+"/.injectived",
		"injectived",
		"file",
		"inj-user",
		"12345678",
		"5d386fbdbf11f1141010f81a46b40f94887367562bd33b452bbaa6ce1cd1381e", // keyring will be used if pk not provided
		false,
	)

	if err != nil {
		panic(err)
	}

	clientCtx, err := chainclient.NewClientContext(
		network.ChainId,
		senderAddress.String(),
		cosmosKeyring,
	)

	if err != nil {
		panic(err)
	}

	clientCtx = clientCtx.WithNodeURI(network.TmEndpoint).WithClient(tmClient)

	chainClient, err := chainclient.NewChainClient(
		clientCtx,
		network,
		common.OptionGasPrices(client.DefaultGasPriceWithDenom),
	)
	if err != nil {
		panic(err)
	}

	exchangeClient, err := exchangeclient.NewExchangeClient(network)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	marketsAssistant, err := chainclient.NewMarketsAssistantInitializedFromChain(ctx, exchangeClient)
	if err != nil {
		panic(err)
	}

	baseToken := marketsAssistant.AllTokens()["INJ"]
	quoteToken := marketsAssistant.AllTokens()["USDC"]
	minPriceTickSize := math.LegacyMustNewDecFromStr("0.01")
	minQuantityTickSize := math.LegacyMustNewDecFromStr("0.01")
	minNotional := math.LegacyMustNewDecFromStr("2")

	chainMinPriceTickSize := minPriceTickSize.Mul(math.LegacyNewDecFromIntWithPrec(math.NewInt(1), int64(quoteToken.Decimals)))
	chainMinPriceTickSize = chainMinPriceTickSize.Quo(math.LegacyNewDecFromIntWithPrec(math.NewInt(1), int64(baseToken.Decimals)))

	chainMinQuantityTickSize := minQuantityTickSize.Mul(math.LegacyNewDecFromIntWithPrec(math.NewInt(1), int64(baseToken.Decimals)))
	chainMinNotional := minNotional.Mul(math.LegacyNewDecFromIntWithPrec(math.NewInt(1), int64(quoteToken.Decimals)))

	msg := &exchangetypes.MsgUpdateSpotMarket{
		Admin:                  senderAddress.String(),
		MarketId:               "0x215970bfdea5c94d8e964a759d3ce6eae1d113900129cc8428267db5ccdb3d1a",
		NewTicker:              "INJ/USDC 2",
		NewMinPriceTickSize:    chainMinPriceTickSize,
		NewMinQuantityTickSize: chainMinQuantityTickSize,
		NewMinNotional:         chainMinNotional,
	}

	// AsyncBroadcastMsg, SyncBroadcastMsg, QueueBroadcastMsg
	response, err := chainClient.AsyncBroadcastMsg(msg)

	if err != nil {
		panic(err)
	}

	str, _ := json.MarshalIndent(response, "", " ")
	fmt.Print(string(str))
}
