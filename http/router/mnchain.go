package router

import (
	"github.com/sixexorg/magnetic-ring/http/resp"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/http/actor"
	"github.com/sixexorg/magnetic-ring/http/req"
	"github.com/sixexorg/magnetic-ring/log"
	"github.com/sixexorg/magnetic-ring/store/mainchain/storages"
	txpool "github.com/sixexorg/magnetic-ring/txpool/mainchain"
)

var (
	blkrg *gin.RouterGroup

	accountTxMap map[string]types.Transactions
)

func TempPutTransaction(tx *types.Transaction) {
	if accountTxMap[tx.TxData.From.ToString()] == nil {
		accountTxMap[tx.TxData.From.ToString()] = make(types.Transactions, 0)
	}
	txns := accountTxMap[tx.TxData.From.ToString()]

	txns = append(txns, tx)

	accountTxMap[tx.TxData.From.ToString()] = txns
}

func TempPutTransactions(txs types.Transactions) {
	for _, tx := range txs {
		TempPutTransaction(tx)
	}
}

func StartRouter() {
	accountTxMap = make(map[string]types.Transactions)

	router := GetRouter()
	blkrg = router.Group("block", func(c *gin.Context) {})

	InitMnBlockChainRouters()

	StartLeagueRouter()
}

func InitMnBlockChainRouters() {
	blkrg.POST("sendtx", sendTx)
	blkrg.POST("gettxs", getTxs)
	blkrg.POST("getbalance", getBalance)
	blkrg.POST("getaccountnonce", getAccountNonce)
	blkrg.POST("getcurrentblockheight", getCurrentBlockHeight)
	blkrg.POST("getcurrentblock", getCurrentBlock)
	blkrg.POST("gettxbytxhash", getTxByHash)
	blkrg.POST("syncorgdata", syncOrgData)
	blkrg.POST("getblockbyheight", getBlockByHeight)
	blkrg.POST("getrefblock", getAccountBlocks)                
}

func sendTx(c *gin.Context) {
	cl := new(req.ReqSendTx)
	err := c.BindJSON(cl)
	if err != nil {
		log.Info("an error occured when bindjson", "err", err)
		return
	}

	log.Info("sendtx", "ReqSendTx", cl)

	tx := &types.Transaction{}
	source := sink.NewZeroCopySource(cl.Raw)
	err = tx.Deserialization(source)
	if err != nil {
		log.Info("transaction Deserialization error", "err", err)
		return
	}

	log.Info("magnetic receive transaction", "tx", tx)

	hx, desc := httpactor.AppendTxToPool(tx)
	desc = "success"

	log.Info("magnetic append transaction to pool", "tx", tx)

	log.Info("watch the hash", "hash", hx)
	resp := txpool.TxResp{
		Hash: hx.String(),
		Desc: desc,
	}

	c.JSON(http.StatusOK, resp)
}

func getTxs(c *gin.Context) {
	cl := new(req.ReqGetTxs)
	err := c.BindJSON(cl)
	if err != nil {
		log.Info("an error occured when bindjson", "error", err)
		return
	}

	log.Info("gettx", "GetTxs", cl)

	txs := accountTxMap[cl.Addr]

	mp := make(map[string]interface{})
	mp["txs"] = txs

	c.JSON(http.StatusOK, mp)
}

func getAccountNonce(c *gin.Context) {
	cl := new(req.ReqAccountNounce)
	err := c.BindJSON(cl)
	if err != nil {
		log.Info("an error occured when bindjson", "error", err)
		return
	}
	mp := make(map[string]interface{})

	accountAddress, err := common.ToAddress(cl.Addr)
	account, err := storages.GetLedgerStore().GetAccount(accountAddress)
	if err != nil {
		c.JSON(http.StatusOK, mp)
		return
	}

	nonce := account.Data.Nonce

	mp["nonce"] = nonce

	c.JSON(http.StatusOK, mp)
}

func getBalance(c *gin.Context) {
	cl := new(req.ReqGetBalance)
	err := c.BindJSON(cl)
	if err != nil {
		log.Info("an error occured when bindjson", "error", err)
		return
	}

	address, err := common.ToAddress(cl.Addr)
	if err != nil {
		mp := make(map[string]interface{})
		mp["crystal"] = 0
		mp["energy"] = 0
		c.JSON(http.StatusOK, mp)
		return
	}

	account, err := storages.GetLedgerStore().GetAccount(address)
	if err != nil {
		panic(err)
	}
	if account == nil {
		mp := make(map[string]interface{})
		mp["crystal"] = 0
		mp["energy"] = 0
		c.JSON(http.StatusOK, mp)
		return
	}

	crystal := account.Data.Balance
	energy := account.Data.EnergyBalance

	mp := make(map[string]interface{})
	mp["crystal"] = crystal
	mp["energy"] = energy

	c.JSON(http.StatusOK, mp)
}

func getCurrentBlockHeight(c *gin.Context) {

	mp := make(map[string]interface{})
	mp["height"] = storages.GetLedgerStore().GetCurrentBlockHeight()

	c.JSON(http.StatusOK, mp)
}

func getCurrentBlock(c *gin.Context) {

	mp := make(map[string]interface{})
	block, _ := storages.GetLedgerStore().GetCurrentBlock()
	mp["block"] = block

	c.JSON(http.StatusOK, mp)
}

func getTxByHash(c *gin.Context) {
	req := new(req.ReqGetTxByHash)
	err := c.BindJSON(req)

	if err != nil {
		log.Info("an error occured when bindjson", "error", err)
		return
	}

	log.Info("get current block height")

	hash, err := common.StringToHash(req.TxHash)
	if err != nil {
		log.Info("an error occured when convert string 2 hash", "error", err)
		return
	}

	tx, _, err := storages.GetLedgerStore().GetTxByHash(hash)

	if err != nil {
		log.Info("an error occured when get tx by hash", "error", err)
		return
	}

	respobj := resp.RespGetTxByHash{
		tx,
	}

	c.JSON(http.StatusOK, respobj)
}

func syncOrgData(c *gin.Context) {
	req := new(req.ReqSyncOrgData)
	err := c.BindJSON(req)

	if err != nil {
		log.Info("an error occured when bindjson", "error", err)
		return
	}

	leagueAddr, err := common.ToAddress(req.LeagueId)
	if err != nil {
		log.Error("convert string to league address error", "err", err)
		return
	}

	httpactor.TellP2pSync(leagueAddr, req.BAdd)

	mp := make(map[string]interface{})
	mp["result"] = "success"

	c.JSON(http.StatusOK, mp)
}

func getBlockByHeight(c *gin.Context) {
	req := new(req.ReqGetBlockByHeight)
	err := c.BindJSON(req)

	if err != nil {
		log.Info("an error occured when bindjson", "error", err)
		return
	}

	mp := make(map[string]interface{})
	block, _ := storages.GetLedgerStore().GetBlockByHeight(req.Height)
	mp["block"] = block

	c.JSON(http.StatusOK, mp)
}

func getAccountBlocks(c *gin.Context) {
	req := new(req.ReqGetRefBlock)
	err := c.BindJSON(req)

	if err != nil {
		log.Info("an error occured when bindjson", "error", err)
		return
	}

	mp := make(map[string]interface{})

	leagueAddr, err := common.ToAddress(req.Account)
	if err != nil {
		log.Error("convert string to league address error", "err", err)
		return
	}

	tres, err := storages.GetLedgerStore().GetAccountRange(req.Start, req.End, leagueAddr)
	if err != nil {
		mp["error"] = err
		c.JSON(http.StatusOK, mp)

		return
	}
	mp["accounts"] = tres
	c.JSON(http.StatusOK, mp)
}

