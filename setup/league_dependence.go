package setup

import (
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/sixexorg/magnetic-ring/account"
	"github.com/sixexorg/magnetic-ring/config"
	"github.com/sixexorg/magnetic-ring/consense/poa"
	"github.com/sixexorg/magnetic-ring/log"
	"github.com/sixexorg/magnetic-ring/meacount"
	"github.com/sixexorg/magnetic-ring/miner"
	"github.com/sixexorg/magnetic-ring/orgcontainer"
)

func initContainers(p2pActor *actor.PID) {
	orgcontainer.InitContainers(p2pActor)
}

func CreateContainer(account account.Account, targetHeight uint64, dbDir string, poaConf *config.CliqueConfig) error {
	node := meacount.GetOwner()
	cns := orgcontainer.GetContainers()
	id, err := cns.NewContainer(dbDir, targetHeight, account, node)
	if err != nil {
		log.Error("create container err", "error", err)
		return err
	}
	log.Info("create container success", "containerId", id)
	for {
		status := cns.GetContainerStatus(id)
		log.Info("collect container status", "status", status)
		if status == orgcontainer.Normal {
			break
		}
		time.Sleep(time.Second * 60)
	}

	container, err := cns.GetContainerById(id)
	if err != nil {
		log.Error("get container by id err", "error", err)
		return err
	}
	log.Info("container success", "leagueId", container.LeagueId().ToString())

	pingpoa := poa.New(poaConf, account, container, container.Ledger())

	poaminer := miner.New(account.Address())

	poaminer.RegisterOrg(container.LeagueId(), container.Ledger(), &config.GlobalConfig.CliqueCfg, pingpoa, container.TxPool()) //config是圈子共识的配置

	poaminer.StartOrg(container.LeagueId())
	return nil
}
