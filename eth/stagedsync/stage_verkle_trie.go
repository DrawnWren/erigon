package stagedsync

import (
	"context"
	"fmt"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/syncreticcapital/erigon/cmd/verkle/verkletrie"
	"github.com/syncreticcapital/erigon/common"
	"github.com/syncreticcapital/erigon/core/rawdb"
	"github.com/syncreticcapital/erigon/eth/stagedsync/stages"
)

func SpawnVerkleTrie(s *StageState, u Unwinder, tx kv.RwTx, cfg TrieCfg, ctx context.Context) (common.Hash, error) {
	var err error
	useExternalTx := tx != nil
	if !useExternalTx {
		tx, err = cfg.db.BeginRw(ctx)
		if err != nil {
			return common.Hash{}, err
		}
		defer tx.Rollback()
	}
	from := uint64(0)
	if s.BlockNumber > 0 {
		from = s.BlockNumber + 1
	}
	to, err := s.ExecutionAt(tx)
	if err != nil {
		return common.Hash{}, err
	}
	verkleWriter := verkletrie.NewVerkleTreeWriter(tx, cfg.tmpDir)
	if err := verkletrie.IncrementAccount(tx, tx, 10, verkleWriter, from, to); err != nil {
		return common.Hash{}, err
	}
	var newRoot common.Hash
	if newRoot, err = verkletrie.IncrementStorage(tx, tx, 10, verkleWriter, from, to); err != nil {
		return common.Hash{}, err
	}
	if cfg.checkRoot {
		header := rawdb.ReadHeaderByNumber(tx, to)
		if header.Root != newRoot {
			return common.Hash{}, fmt.Errorf("invalid verkle root, header has %x, computed: %x", header.Root, newRoot)
		}
	}
	if err := s.Update(tx, to); err != nil {
		return common.Hash{}, err
	}
	if err := stages.SaveStageProgress(tx, stages.VerkleTrie, to); err != nil {
		return common.Hash{}, err
	}
	if !useExternalTx {
		return newRoot, tx.Commit()
	}
	return newRoot, nil
}

func UnwindVerkleTrie(u *UnwindState, s *StageState, tx kv.RwTx, cfg TrieCfg, ctx context.Context) (err error) {
	useExternalTx := tx != nil
	if !useExternalTx {
		tx, err = cfg.db.BeginRw(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback()
	}
	from := u.UnwindPoint + 1
	to, err := s.ExecutionAt(tx)
	if err != nil {
		return err
	}
	verkleWriter := verkletrie.NewVerkleTreeWriter(tx, cfg.tmpDir)
	if err := verkletrie.IncrementAccount(tx, tx, 10, verkleWriter, from, to); err != nil {
		return err
	}
	if _, err = verkletrie.IncrementStorage(tx, tx, 10, verkleWriter, from, to); err != nil {
		return err
	}
	if err := s.Update(tx, from); err != nil {
		return err
	}
	if err := stages.SaveStageProgress(tx, stages.VerkleTrie, from); err != nil {
		return err
	}
	if !useExternalTx {
		return tx.Commit()
	}
	return nil
}

func PruneVerkleTries(s *PruneState, tx kv.RwTx, cfg TrieCfg, ctx context.Context) (err error) {
	useExternalTx := tx != nil
	if !useExternalTx {
		tx, err = cfg.db.BeginRw(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback()
	}
	s.Done(tx)

	if !useExternalTx {
		if err = tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
