package health

import (
	"context"

	"github.com/syncreticcapital/erigon/common/hexutil"
	"github.com/syncreticcapital/erigon/rpc"
)

type NetAPI interface {
	PeerCount(_ context.Context) (hexutil.Uint, error)
}

type EthAPI interface {
	GetBlockByNumber(_ context.Context, number rpc.BlockNumber, fullTx bool) (map[string]interface{}, error)
	Syncing(ctx context.Context) (interface{}, error)
}
