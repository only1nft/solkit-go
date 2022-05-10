package genesysgo_test

import (
	"context"
	"os"
	"testing"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/only1nft/solkit-go/genesysgo"
	"github.com/test-go/testify/assert"
)

func TestNewRPCClientDevNetWithoutToken(t *testing.T) {
	conn := genesysgo.NewRPCClient(rpc.DevNet.RPC, "")
	res, err := conn.GetLatestBlockhash(context.TODO(), rpc.CommitmentConfirmed)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestNewRPCClientMainNetBetaWithEmptyToken(t *testing.T) {
	conn := genesysgo.NewRPCClient(genesysgo.MainNetBeta.RPC, "")
	res, err := conn.GetLatestBlockhash(context.TODO(), rpc.CommitmentConfirmed)
	assert.Error(t, err, "Unauthorize")
	assert.Nil(t, res)
}

func TestNewRPCClientMainNetBetaWithToken(t *testing.T) {
	conn := genesysgo.NewRPCClient(genesysgo.MainNetBeta.RPC, os.Getenv("GENESYSGO_TOKEN"))
	res, err := conn.GetLatestBlockhash(context.TODO(), rpc.CommitmentConfirmed)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}
