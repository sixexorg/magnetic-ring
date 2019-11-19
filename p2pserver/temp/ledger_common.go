package temp

import (
	"fmt"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/log"
	"github.com/sixexorg/magnetic-ring/orgcontainer"
	"github.com/sixexorg/magnetic-ring/store/orgchain/storages"
)

func GetContainerByOrgID(orgid common.Address) (*orgcontainer.Container,error) {
	gcontainer := orgcontainer.GetContainers()
	if gcontainer == nil {
		return nil,fmt.Errorf("GetContainerByOrgID gcontainer is nil")
	}

	container,err := gcontainer.GetContainer(orgid)
	if err != nil {
		return nil,err
	}
	if container == nil {
		return nil,fmt.Errorf("GetContainerByOrgID container is nil")
	}
	return container,err
}

//

func GetLedger(orgid common.Address,flags ... string) *storages.LedgerStoreImp{
	container,err := GetContainerByOrgID(orgid)
	if err != nil {
		log.Error("GetLedger","orgid",orgid.ToString(),"err",err,"flags",flags)
		return nil
	}
	return container.Ledger()
}