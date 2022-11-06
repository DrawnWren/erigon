package ethconsensusconfig

import (
	"path/filepath"

	"github.com/ledgerwatch/erigon-lib/kv"

	"github.com/davecgh/go-spew/spew"
	"github.com/syncreticcapital/erigon/consensus"
	"github.com/syncreticcapital/erigon/consensus/aura"
	"github.com/syncreticcapital/erigon/consensus/aura/consensusconfig"
	"github.com/syncreticcapital/erigon/consensus/bor"
	"github.com/syncreticcapital/erigon/consensus/clique"
	"github.com/syncreticcapital/erigon/consensus/db"
	"github.com/syncreticcapital/erigon/consensus/ethash"
	"github.com/syncreticcapital/erigon/consensus/parlia"
	"github.com/syncreticcapital/erigon/consensus/serenity"
	"github.com/syncreticcapital/erigon/params"
	"github.com/syncreticcapital/erigon/turbo/snapshotsync"
	"github.com/ledgerwatch/log/v3"
)

func CreateConsensusEngine(chainConfig *params.ChainConfig, logger log.Logger, config interface{}, notify []string, noverify bool, HeimdallURL string, WithoutHeimdall bool, datadir string, snapshots *snapshotsync.RoSnapshots, readonly bool, chainDb ...kv.RwDB) consensus.Engine {
	var eng consensus.Engine

	switch consensusCfg := config.(type) {
	case *ethash.Config:
		switch consensusCfg.PowMode {
		case ethash.ModeFake:
			log.Warn("Ethash used in fake mode")
			eng = ethash.NewFaker()
		case ethash.ModeTest:
			log.Warn("Ethash used in test mode")
			eng = ethash.NewTester(nil, noverify)
		case ethash.ModeShared:
			log.Warn("Ethash used in shared mode")
			eng = ethash.NewShared()
		default:
			eng = ethash.New(ethash.Config{
				CachesInMem:      consensusCfg.CachesInMem,
				CachesLockMmap:   consensusCfg.CachesLockMmap,
				DatasetDir:       consensusCfg.DatasetDir,
				DatasetsInMem:    consensusCfg.DatasetsInMem,
				DatasetsOnDisk:   consensusCfg.DatasetsOnDisk,
				DatasetsLockMmap: consensusCfg.DatasetsLockMmap,
			}, notify, noverify)
		}
	case *params.ConsensusSnapshotConfig:
		if chainConfig.Clique != nil {
			if consensusCfg.DBPath == "" {
				consensusCfg.DBPath = filepath.Join(datadir, "clique", "db")
			}
			eng = clique.New(chainConfig, consensusCfg, db.OpenDatabase(consensusCfg.DBPath, logger, consensusCfg.InMemory, readonly))
		}
	case *params.AuRaConfig:
		if chainConfig.Aura != nil {
			if consensusCfg.DBPath == "" {
				consensusCfg.DBPath = filepath.Join(datadir, "aura")
			}
			var err error
			eng, err = aura.NewAuRa(chainConfig.Aura, db.OpenDatabase(consensusCfg.DBPath, logger, consensusCfg.InMemory, readonly), chainConfig.Aura.Etherbase, consensusconfig.GetConfigByChain(chainConfig.ChainName))
			if err != nil {
				panic(err)
			}
		}
	case *params.ParliaConfig:
		if chainConfig.Parlia != nil {
			if consensusCfg.DBPath == "" {
				consensusCfg.DBPath = filepath.Join(datadir, "parlia")
			}
			eng = parlia.New(chainConfig, db.OpenDatabase(consensusCfg.DBPath, logger, consensusCfg.InMemory, readonly), snapshots, chainDb[0])
		}
	case *params.BorConfig:
		if chainConfig.Bor != nil {
			borDbPath := filepath.Join(datadir, "bor") // bor consensus path: datadir/bor
			eng = bor.New(chainConfig, db.OpenDatabase(borDbPath, logger, false, readonly), HeimdallURL, WithoutHeimdall)
		}
	}

	if eng == nil {
		panic("unknown config" + spew.Sdump(config))
	}

	if chainConfig.TerminalTotalDifficulty == nil {
		return eng
	} else {
		return serenity.New(eng) // the Merge
	}
}
