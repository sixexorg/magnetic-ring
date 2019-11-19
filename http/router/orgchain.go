package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sixexorg/magnetic-ring/central/entity"
	"github.com/sixexorg/magnetic-ring/central/mongo/service/impl"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
	httpactor "github.com/sixexorg/magnetic-ring/http/actor"
	"github.com/sixexorg/magnetic-ring/http/req"
	"github.com/sixexorg/magnetic-ring/log"
	"github.com/sixexorg/magnetic-ring/orgcontainer"
	"github.com/sixexorg/magnetic-ring/store/orgchain/validation"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

var (
	leagueGroup *gin.RouterGroup
)

func StartLeagueRouter() {

	router := GetRouter()
	leagueGroup = router.Group("orgapi", func(c *gin.Context) {})

	InitOrgRouters()
}

func InitOrgRouters() {
	leagueGroup.POST("getleaguenonce", getLeagueNonce)
	leagueGroup.POST("sendsubtx", sendSubTx)
	leagueGroup.POST("getorgbalance", getSubBalance)
	leagueGroup.POST("getorgheight", getSubCurrentBlockHeight)
	leagueGroup.POST("getorgblock", getSubBlock)
	leagueGroup.POST("getorgtxbyhash", getSubTxByHash)
	leagueGroup.POST("getorginfo", getLeague)
	leagueGroup.POST("getorgvotes", getVoteList)
	leagueGroup.POST("getorgmembers", getLeagueMemberList)
	leagueGroup.POST("getorgdetail", getLeagueDetail)
	leagueGroup.POST("ungetbonus", unGetBonus)
	leagueGroup.POST("leagueasset", leagueAsset)
	leagueGroup.POST("getorgvotedetail", getVoteDetail)
	leagueGroup.POST("getchatlist", getChatRecordList)
	leagueGroup.POST("getransferlist", getTransferList)
	leagueGroup.POST("getvoteunitlist", getVoteUnitList)
}

func sendSubTx(c *gin.Context) {
	cl := new(req.ReqSendTx)
	err := c.BindJSON(cl)
	bsm := bson.M{}
	if err != nil {
		log.Info("an error occured when bindjson", "err", err)
		bsm["errormsg"] = err.Error()
		c.JSON(http.StatusOK, bsm)
		return
	}

	log.Info("sendtx", "ReqSendTx", cl)

	tx := &types.Transaction{}
	source := sink.NewZeroCopySource(cl.Raw)
	err = tx.Deserialization(source)
	if err != nil {
		log.Info("transaction Deserialization error", "err", err)
		bsm["errormsg"] = err.Error()
		c.JSON(http.StatusOK, bsm)
		return
	}

	log.Info("magnetic receive transaction", "tx", tx)

	leagueAddress, err := common.ToAddress(cl.LeagueId)
	if err != nil {
		log.Error("toaddress errorrrr", "error", err)
		bsm["errormsg"] = err.Error()
		c.JSON(http.StatusOK, bsm)
		return
	}

	hx, desc := httpactor.AppendTxToSubPool(tx, leagueAddress)
	desc = "success"

	log.Info("magnetic append transaction to pool", "tx", tx)

	log.Info("watch the hash", "hash", hx)

	bsm["errormsg"] = ""
	bsm["hash"] = hx
	bsm["desc"] = desc

	c.JSON(http.StatusOK, bsm)
}

func RegistGetSubBalance() {
	leagueGroup.POST("getorgbalance")
}

func getSubBalance(c *gin.Context) {
	cl := new(req.ReqGetBalance)
	err := c.BindJSON(cl)
	if err != nil {
		log.Info("an error occured when bindjson", "error", err)
		return
	}

	address, err := common.ToAddress(cl.OrgId)
	if err != nil {
		mp := make(map[string]interface{})
		mp["ut"] = 0
		mp["energy"] = 0
		c.JSON(http.StatusOK, mp)
		return
	}

	ctner, err := orgcontainer.GetContainers().GetContainer(address)
	if err != nil {
		panic(err)
	}

	ldgr := ctner.Ledger()

	mp := make(map[string]interface{})
	account, err := ldgr.GetPrevAccount4V(ldgr.GetCurrentBlockHeight(), address, common.Address{})
	if err != nil {
		mp["error"] = err.Error()
		c.JSON(http.StatusOK, mp)
		return
	}

	if account == nil {
		mp["ut"] = 0
		mp["energy"] = 0
		c.JSON(http.StatusOK, mp)
		return
	}

	ut := account.Balance()
	energy := account.Energy()

	mp["ut"] = ut
	mp["energy"] = energy

	c.JSON(http.StatusOK, mp)
}

func getSubCurrentBlockHeight(c *gin.Context) {
	cl := new(req.ReqGetBalance)
	err := c.BindJSON(cl)
	if err != nil {
		log.Info("an error occured when bindjson", "error", err)
		return
	}

	address, err := common.ToAddress(cl.Addr)
	if err != nil {
		mp := make(map[string]interface{})
		mp["error"] = err.Error()
		c.JSON(http.StatusOK, mp)
		return
	}

	ctner, err := orgcontainer.GetContainers().GetContainer(address)
	if err != nil {
		panic(err)
	}

	ldgr := ctner.Ledger()

	mp := make(map[string]interface{})
	height := ldgr.GetCurrentBlockHeight()

	mp["height"] = height

	c.JSON(http.StatusOK, mp)
}

func getSubBlock(c *gin.Context) {
	cl := new(req.ReqGetBlockByHeight)
	err := c.BindJSON(cl)
	if err != nil {
		log.Info("an error occured when bindjson", "error", err)
		return
	}

	address, err := common.ToAddress(cl.OrgId)
	if err != nil {
		mp := make(map[string]interface{})
		mp["error"] = err.Error()
		c.JSON(http.StatusOK, mp)
		return
	}

	ctner, err := orgcontainer.GetContainers().GetContainer(address)
	if err != nil {
		panic(err)
	}

	ldgr := ctner.Ledger()

	mp := make(map[string]interface{})
	block, err := ldgr.GetBlockByHeight(cl.Height)
	if err != nil {
		mp["error"] = err.Error()
		c.JSON(http.StatusOK, mp)
		return
	}

	mp["block"] = block

	c.JSON(http.StatusOK, mp)
}

func getSubTxByHash(c *gin.Context) {
	cl := new(req.ReqGetTxByHash)
	err := c.BindJSON(cl)
	if err != nil {
		log.Info("an error occured when bindjson", "error", err)
		return
	}

	address, err := common.ToAddress(cl.LeagueId)
	if err != nil {
		mp := make(map[string]interface{})
		mp["error"] = err.Error()
		c.JSON(http.StatusOK, mp)
		return
	}

	ctner, err := orgcontainer.GetContainers().GetContainer(address)
	if err != nil {
		panic(err)
	}

	ldgr := ctner.Ledger()

	hash, err := common.StringToHash(cl.TxHash)
	if err != nil {
		log.Info("an error occured when convert string 2 hash", "error", err)
		return
	}

	mp := make(map[string]interface{})
	block, height, err := ldgr.GetTransaction(hash, address)
	if err != nil {
		mp["error"] = err.Error()
		c.JSON(http.StatusOK, mp)
		return
	}

	mp["tx"] = block
	mp["height"] = height

	c.JSON(http.StatusOK, mp)
}

func getLeague(c *gin.Context) {
	cl := new(req.ReqOrgInfo)
	err := c.BindJSON(cl)
	mp := make(map[string]interface{})
	if err != nil {
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	address, err := common.ToAddress(cl.OrgId)
	if err != nil {
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	ctner, err := orgcontainer.GetContainers().GetContainer(address)
	if err != nil {
		mp := make(map[string]interface{})
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	mp["orgid"] = ctner.LeagueId().ToString()
	c.JSON(http.StatusOK, mp)
}

/*****************************************************************/

func getLeagueDetail(c *gin.Context) {
	cl := new(req.ReqOrgInfo)
	err := c.BindJSON(cl)
	mp := make(map[string]interface{})
	if err != nil {
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	mp["ut"] = 300000
	mp["totalboxes"] = 1000
	mp["orgSymbol"] = "CHINA"
	mp["utrate"] = 1000
	mp["userauth"] = true
	mp["orgaddress"] = "ct4x6b5nK4JVTFFHRW947RAECX2xT4E4Fk4"

	c.JSON(http.StatusOK, mp)
}

func getLeagueMemberList(c *gin.Context) {
	cl := new(req.ReqGetMemberList)
	err := c.BindJSON(cl)
	mp := make(map[string]interface{})
	if err != nil {
		log.Error("impl.NewMemberImpl(nil,cl.OrgId).GetMebers()", "error", err)
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	total, mbs, err := impl.NewMemberImpl(nil, cl.OrgId).GetMebers(cl.Page, cl.PageSize)
	if err != nil {
		log.Error("impl.NewMemberImpl(nil,cl.OrgId).GetMebers()", "error", err)
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	mbslit := mbs.([]*entity.Member)

	memberlist := make([]string, 0) //{"ct1x6b5nK4JVTFFHRW947RAECX2xT4E4Fk0", "ct1x6b5nK4JVTFFHRW947RAECX2xT4E4Fk1", "ct1x6b5nK4JVTFFHRW947RAECX2xT4E4Fk2", "ct1x6b5nK4JVTFFHRW947RAECX2xT4E4Fk4", "ct1x6b5nK4JVTFFHRW947RAECX2xT4E4Fk4", "ct1x6b5nK4JVTFFHRW947RAECX2xT4E4Fk5"}

	for _, tmp := range mbslit {
		//tmp := mem.(*entity.Member)
		memberlist = append(memberlist, tmp.Addr)
	}

	mp["total"] = total
	mp["memberlist"] = memberlist

	c.JSON(http.StatusOK, mp)
}

func getVoteList(c *gin.Context) {
	cl := new(req.ReqGetVoteList)
	err := c.BindJSON(cl)
	mp := make(map[string]interface{})
	if err != nil {
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	address, err := common.ToAddress(cl.OrgId)
	if err != nil {
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	ctner, err := orgcontainer.GetContainers().GetContainer(address)
	if err != nil {
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	ldgr := ctner.Ledger()

	total, list, err := impl.NewVoteDetailImpl(nil, cl.OrgId).GetVoteList(cl.Page, cl.PageSize, ldgr.GetCurrentBlockHeight(), cl.VoteState)
	if err != nil {
		log.Error("impl.NewMemberImpl(nil,cl.OrgId).GetVoteList()", "error", err)
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	mp["votelist"] = list
	mp["total"] = total

	c.JSON(http.StatusOK, mp)
}

func getChatRecordList(c *gin.Context) {
	cl := new(req.ReqGetVoteList)
	err := c.BindJSON(cl)
	mp := make(map[string]interface{})
	if err != nil {
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	//addr, _ := common.ToAddress("ct1x6b5nK4JVTFFHRW947RAECX2xT4E4Fk0")
	//hs, _ := common.StringToHash("53ee84c855b17595543cd26d41ff40504f2cbdd5296c537aac8b94f716053331")

	//vlist := []*entity.VoteUnit{
	//	entity.NewVoteUnit(types.Join, addr, 8, hs),
	//	entity.NewVoteUnit(types.Leave, addr, 1608, hs),
	//	entity.NewVoteUnit(types.RaiseUT, addr, 3608, hs),
	//}

	total, list, err := impl.NewOrgChatMsgImpl(nil, cl.OrgId).GetChatList(cl.Page, cl.PageSize)

	mp["chatrecordlist"] = list
	mp["total"] = total

	c.JSON(http.StatusOK, mp)
}

func unGetBonus(c *gin.Context) {
	cl := new(req.ReqUngetBonus)
	err := c.BindJSON(cl)
	mp := make(map[string]interface{})
	if err != nil {
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	address, err := common.ToAddress(cl.OrgId)
	if err != nil {
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	ctner, err := orgcontainer.GetContainers().GetContainer(address)
	if err != nil {
		mp := make(map[string]interface{})
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	accountAddress, err := common.ToAddress(cl.Account)
	if err != nil {
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	energy, _, err := validation.CalculateBonus(ctner.Ledger(), accountAddress, address)
	if err != nil {
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	mp["energy"] = energy.Uint64()

	c.JSON(http.StatusOK, mp)
}

func leagueAsset(c *gin.Context) {
	cl := new(req.ReqGetOrgAsset)
	err := c.BindJSON(cl)
	mp := make(map[string]interface{})
	if err != nil {
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	address, err := common.ToAddress(cl.OrgId)
	if err != nil {
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	ctner, err := orgcontainer.GetContainers().GetContainer(address)
	if err != nil {
		mp := make(map[string]interface{})
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	acctaddress, err := common.ToAddress(cl.Account)
	if err != nil {
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	stt, err := ctner.Ledger().GetPrevAccount4V(cl.Height, acctaddress, address)
	if err != nil {
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	mp["ut"] = stt.Balance()
	mp["energy"] = stt.Energy()

	c.JSON(http.StatusOK, mp)
}

func getVoteDetail(c *gin.Context) {
	cl := new(req.ReqVoteDetail)
	err := c.BindJSON(cl)
	mp := make(map[string]interface{})
	if err != nil {
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	address, err := common.ToAddress(cl.OrgId)
	if err != nil {
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	ctner, err := orgcontainer.GetContainers().GetContainer(address)
	if err != nil {
		mp := make(map[string]interface{})
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	hsh, err := common.StringToHash(cl.VoteId)
	if err != nil {
		mp := make(map[string]interface{})
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	detail, err := impl.NewVoteDetailImpl(nil, cl.OrgId).GetLeagueVoteDetail(cl.OrgId, cl.VoteId)
	if err != nil {
		mp := make(map[string]interface{})
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	vstt, err := ctner.Ledger().GetVoteState(hsh, cl.Height)
	if err != nil {
		mp := make(map[string]interface{})
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	uttotal := ctner.Ledger().GetUTByHeight(cl.Height, common.Address{})

	agreeper, _ := common.CalcPer(vstt.Agree, uttotal)
	agaitper, _ := common.CalcPer(vstt.Against, uttotal)
	giveuper, _ := common.CalcPer(vstt.Abstention, uttotal)

	result := vstt.GetResult(uttotal, 67)

	acctaddress, err := common.ToAddress(cl.Account)
	if err != nil {
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	accountState, err := ctner.Ledger().GetPrevAccount4V(cl.Height, acctaddress, address)
	if err != nil {
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	myut := accountState.Balance()

	mystat := impl.NewVoteStatisticsImpl(nil, cl.OrgId).IsAccountVoted(cl.Account)

	vlist := entity.VoteDetail{
		MetaBox:    detail.MetaBox,
		Hoster:     detail.Hoster,
		AgreePer:   agreeper,
		AgainstPer: agaitper,
		GiveUpPer:  giveuper,
		UtBefore:   detail.UtBefore,
		MyUt:       myut.Uint64(),
		Msg:        detail.Msg,
		State:      result,
		EndHeight:  detail.EndHeight,
		IfIVoted:   mystat,
	}

	mp["votedetail"] = vlist

	c.JSON(http.StatusOK, mp)
}

func getTransferList(c *gin.Context) {
	cl := new(req.ReqGetVoteList)
	err := c.BindJSON(cl)
	mp := make(map[string]interface{})
	if err != nil {
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	total, list, err := impl.NewOrgTxnsImpl(nil, cl.OrgId).GetTransferTxList(cl.Account, cl.Page, cl.PageSize)

	mp["transferlist"] = list
	mp["total"] = total

	c.JSON(http.StatusOK, mp)
}

func getLeagueNonce(c *gin.Context) {
	cl := new(req.ReqUngetBonus)
	err := c.BindJSON(cl)
	if err != nil {
		log.Info("an error occured when bindjson", "error", err)
		return
	}

	leagueAddress, err := common.ToAddress(cl.OrgId)
	if err != nil {
		mp := make(map[string]interface{})
		mp["error"] = err.Error()
		c.JSON(http.StatusOK, mp)
		return
	}

	accountAddress, err := common.ToAddress(cl.Account)
	if err != nil {
		mp := make(map[string]interface{})
		mp["error"] = err.Error()
		c.JSON(http.StatusOK, mp)
		return
	}

	empty := common.Address{}
	if leagueAddress == empty {
		mp := make(map[string]interface{})
		mp["error"] = err.Error()
		c.JSON(http.StatusOK, mp)
		return
	}

	ctner, err := orgcontainer.GetContainers().GetContainer(leagueAddress)
	if err != nil {
		panic(err)
	}

	ldgr := ctner.Ledger()

	acctstate, err := ldgr.GetPrevAccount4V(ldgr.GetCurrentBlockHeight(), accountAddress, leagueAddress)
	if err != nil {
		mp := make(map[string]interface{})
		mp["error"] = err.Error()
		c.JSON(http.StatusOK, mp)
		return
	}

	mp := make(map[string]interface{})
	mp["nonce"] = acctstate.Nonce()

	c.JSON(http.StatusOK, mp)
}

func getVoteUnitList(c *gin.Context) {
	cl := new(req.ReqGetVoteUnitList)
	err := c.BindJSON(cl)
	mp := make(map[string]interface{})
	if err != nil {
		mp["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, mp)
		return
	}

	total, list, err := impl.NewVoteStatisticsImpl(nil, cl.OrgId).FindVoteUnitList(cl.OrgId, cl.VoteId, cl.Page, cl.PageSize)

	mp["voteunitlist"] = list
	mp["total"] = total

	c.JSON(http.StatusOK, mp)
}
