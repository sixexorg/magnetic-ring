package orgchain

import (
	"github.com/sixexorg/magnetic-ring/store/orgchain/storages"
)

type radarAdapter struct {
	accRootSotre *storages.AccountRootStore
	blockStore   *storages.BlockStore
	accountStore *storages.AccountStore
}
