package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	txpool "github.com/sixexorg/magnetic-ring/txpool/mainchain"
	"io/ioutil"
	"net/http"
	"time"
)

func HttpSend(fullurl string, request interface{}) (string, error) {
	fmt.Printf("test req fullurl=%s\n",fullurl)
	reqbuf, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	buffer := bytes.NewBuffer(reqbuf)
	req, err := http.NewRequest("POST", fullurl, buffer)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := http.Client{}
	client.Timeout = time.Second * 3
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	respobj := new(txpool.TxResp)
	err = json.Unmarshal(buf, respobj)
	if err != nil {
		return "", err
	}
	return string(buf), nil

}
