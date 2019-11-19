package walletutil

import (
	"fmt"
	"github.com/sixexorg/magnetic-ring/common"
	httpcli "github.com/sixexorg/magnetic-ring/http/cli"
	"github.com/sixexorg/magnetic-ring/http/req"
)

func PushTransaction(url string, txhex string) string {
	ra,err := common.Hex2Bytes(txhex)
	if err != nil {
		return newErrorResp(err.Error())
	}
	req := req.ReqSendTx{
		Raw: ra,
	}
	fullurl := fmt.Sprintf("%s/block/sendtx", url)
	resp, err := httpcli.HttpSend(fullurl, req)
	if err != nil {
		return newErrorResp(err.Error())
	}
	return newResp(fmt.Sprintf("%s", resp))
}


func GetCurrentBlockHeigt(url string) string {
	fullurl := fmt.Sprintf("%s/block/getcurrentblock", url)
	resp, err := httpcli.HttpSend(fullurl, nil)
	if err != nil {
		return newErrorResp(err.Error())
	}
	return newResp(fmt.Sprintf("%s", resp))
}


func Getaccountnonce(url,account string) string {
	fullurl := fmt.Sprintf("%s/block/getaccountnonce", url)

	req := req.ReqAccountNounce{account}

	resp, err := httpcli.HttpSend(fullurl, req)
	rt := ""
	if err != nil {
		rt = newErrorResp(err.Error())
	}
	rt = newResp(fmt.Sprintf("%s", resp))
	return rt
}

func Getblockbyheight(url string,height int64,orgId string) string {
	fullurl := fmt.Sprintf("%s/block/getblockbyheight", url)

	req := req.ReqGetBlockByHeight{uint64(height),orgId}

	resp, err := httpcli.HttpSend(fullurl, req)
	if err != nil {
		return newErrorResp(err.Error())
	}
	return newResp(fmt.Sprintf("%s", resp))
}


func GetAccountBalance(url,account,orgid string) string {
	fullurl := fmt.Sprintf("%s/block/getbalance",url)

	req := req.ReqGetBalance{orgid,account}

	resp, err := httpcli.HttpSend(fullurl, req)
	if err != nil {
		return newErrorResp(err.Error())
	}
	return newResp(fmt.Sprintf("%s", resp))
}


func GetAccountRefBlockHeights(url,account string,start,end int64) string {
	fullurl := fmt.Sprintf("%s/block/getrefblock", url)

	req := req.ReqGetRefBlock{
		uint64(start),
		uint64(end),
		account,
	}

	resp, err := httpcli.HttpSend(fullurl, req)
	if err != nil {
		return newErrorResp(err.Error())
	}
	return newResp(fmt.Sprintf("%s", resp))
}

func GetOrgInfo(url,orgid string) string {
	fullurl := fmt.Sprintf("%s/block/getorginfo", url)

	req := req.ReqOrgInfo{
		orgid,
	}

	resp, err := httpcli.HttpSend(fullurl, req)
	if err != nil {
		return newErrorResp(err.Error())
	}
	return newResp(fmt.Sprintf("%s", resp))
}