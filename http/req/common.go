package req

type ReqSendTx struct {
	LeagueId string `json:"league_id"`
	Raw      []byte `json:"raw"`
}

type ReqGetTxs struct {
	Addr string
}

type ReqGetBalance struct {
	OrgId string `json:"org_id"`
	Addr   string `json:"addr"`
	//Height uint64 `json:"height"`
}

type ReqAccountNounce struct {
	Addr string
}

type ReqGetTxByHash struct {
	LeagueId string `json:"league_id"`
	TxHash   string `json:"tx_hash"`
}

type ReqGetBlockByHeight struct {
	Height uint64 `json:"height"`
	OrgId string `json:"org_id"`
}

type ReqSyncOrgData struct {
	LeagueId string
	BAdd     bool
}

type ReqGetRefBlock struct {
	Start   uint64 `json:"start"`
	End     uint64 `json:"end"`
	Account string `json:"account"`
}

type ReqOrgInfo struct {
	OrgId string `json:"org_id"`
}

type ReqGetVoteList struct {
	ReqGetMemberList
	VoteState uint8 `json:"vote_state"`
}

type ReqUngetBonus struct {
	OrgId   string `json:"org_id"`
	Account string `json:"account"`
}

type ReqPageBase struct {
	Page int `json:"page"`
	PageSize int `json:"page_size"`
}

type ReqGetAssetTxList struct {
	ReqGetMemberList
	AssetType uint8 `json:"asset_type"`
}

type ReqGetMemberList struct {
	ReqPageBase
	OrgId   string `json:"org_id"`
	Account string `json:"account"`
}

type ReqVoteDetail struct {
	ReqUngetBonus
	VoteId string `json:"vote_id"`
	Height uint64 `json:"height"`
}

type ReqGetOrgAsset struct {
	OrgId   string `json:"org_id"`
	Height uint64 `json:"height"`
	Account string `json:"account"`
}

type ReqGetVoteUnitList struct {
	ReqGetMemberList
	VoteId string
}