package main

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	"github.com/metachris/flashbotsrpc"
	"log"
	"math/big"
	"os"
)

type Bundle struct {
	bundleId string
	rawTxs   []string
}
type Bot struct {
	Client     *ethclient.Client
	SigningKey *ecdsa.PrivateKey
	PrivateKey *ecdsa.PrivateKey
	Address    *common.Address
}

var err error

func toGwei(wei *big.Int) *big.Int {

	return new(big.Int).Div(wei, big.NewInt(1000000000))
}

func NewBot(signingKey string, privateKey string, providerURL string) (*Bot, error) {
	var err error
	bot := Bot{}

	bot.Client, err = ethclient.Dial(providerURL)
	if err != nil {
		return nil, err
	}

	bot.SigningKey, err = crypto.HexToECDSA(signingKey)
	if err != nil {
		return nil, err
	}

	bot.PrivateKey, bot.Address, err = wallet(privateKey)
	if err != nil {
		return nil, err
	}

	return &bot, nil
}

func send() {
	bunleTxs := []string{
		"0x02f87201028512a05f200085174876e8008252089425df6da2f4e5c178ddff45038378c0b08e0bce54865af3107a400080c001a04e7f16419eb1185c95994a06898c28ff5e6e2c5d3787f1b880c58ec67cb4a8d2a05db305b0a34d012d21dceee76bfa0e32679077697715c80fe69103177a4e0b1c",
		"0x02f868010380808252089425df6da2f4e5c178ddff45038378c0b08e0bce54865af3107a400080c001a028af5ce74851a2d7f00692a8d5f95a14e79026adef87093edd052abb38228865a039d90a2df176d0d1193dcb5d2e50ca9525ba974a80b6dbdf64d0f75c107a44b5",
	}
	// Load config from env
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	privateKey := os.Getenv("BOT_PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("BOT_PRIVATE_KEY not set")
	}

	signingKey := os.Getenv("FLASHBOTS_SIGNING_KEY")
	if signingKey == "" {
		log.Fatal("FLASHBOTS_SIGNING_KEY not set")
	}

	providerURL := os.Getenv("PROVIDER_URL")
	if providerURL == "" {
		log.Fatal("PROVIDER_URL not set")
	}
	bot, err := NewBot(signingKey, privateKey, providerURL)
	if err != nil {
		fmt.Printf("Failed to initialize bot: %s\n", err)
		return
	}

	rpc := flashbotsrpc.New("https://relay.flashbots.net")

	// Simulate bundle
	opts := flashbotsrpc.FlashbotsCallBundleParam{
		Txs:              bunleTxs,
		BlockNumber:      fmt.Sprintf("0x%x", 13281018),
		StateBlockNumber: "latest",
	}

	result, err := rpc.FlashbotsCallBundle(bot.PrivateKey, opts)
	if err != nil {
		if errors.Is(err, flashbotsrpc.ErrRelayErrorResponse) { // user/tx error, rather than JSON or network error
			fmt.Println(err.Error())
		} else {
			fmt.Printf("error: %+v\n", err)
		}
		return
	}

	// Print result
	fmt.Printf("%+v\n", result)

	// Send bundle

	rpc.Debug = true

	sendBundleArgs := flashbotsrpc.FlashbotsSendBundleRequest{
		Txs:         bunleTxs,
		BlockNumber: fmt.Sprintf("0x%x", 13281018),
	}

	resultSend, err := rpc.FlashbotsSendBundle(bot.PrivateKey, sendBundleArgs)
	if err != nil {
		if errors.Is(err, flashbotsrpc.ErrRelayErrorResponse) {
			// ErrRelayErrorResponse means it's a standard Flashbots relay error response, so probably a user error, rather than JSON or network error
			fmt.Println(err.Error())
		} else {
			fmt.Printf("error: %+v\n", err)
		}
		return
	}

	// Print result
	fmt.Printf("%+v\n", resultSend)

}

func wallet(private string) (*ecdsa.PrivateKey, *common.Address, error) {

	privateKey, err := crypto.HexToECDSA(private)
	if err != nil {
		return nil, nil, err
	}
	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	return privateKey, &fromAddress, nil
}

func main() {
	fmt.Println("Starting up")

	send()
}
