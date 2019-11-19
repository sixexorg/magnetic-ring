package storages

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

var (
	stars = []string{
		//"04dc2c38fd4985a30f31fe9ab0e1d6ffa85d33d5c8f0a5ff38c45534e8f4ffe751055e6f1bbec984839dd772a68794722683ed12994feb84a13b8086162591418f",
		//"044a147deabaa89e15aab6f586ce8c9b68bf11043ca83387b125d82489252d94e858f7b43a43000d7a949ff1bd8742fcad57434d08b2455bfc5f681dd0cf0a32f6",
		//"0405269acdc54c24220f67911a3ac709b129ff2454717875b789288954bb2e6afaed4d59ff0f4c4d14afff9f6f4e1ffb71a1e5457fa2ca7440a020218559ab7f3f",
		//"04dc2c38fd4985a30f31fe9ab0e1d6ffa85d33d5c8f0a5ff38c45534e8f4ffe751055e6f1bbec984839dd772a68794722683ed12994feb84a13b8086162591418f",
		//"044a147deabaa89e15aab6f586ce8c9b68bf11043ca83387b125d82489252d94e858f7b43a43000d7a949ff1bd8742fcad57434d08b2455bfc5f681dd0cf0a32f6",
		//"0405269acdc54c24220f67911a3ac709b129ff2454717875b789288954bb2e6afaed4d59ff0f4c4d14afff9f6f4e1ffb71a1e5457fa2ca7440a020218559ab7f3f",
		"04dc2c38fd4985a30f31fe9ab0e1d6ffa85d33d5c8f0a5ff38c45534e8f4ffe751055e6f1bbec984839dd772a68794722683ed12994feb84a13b8086162591418f",
		"044a147deabaa89e15aab6f586ce8c9b68bf11043ca83387b125d82489252d94e858f7b43a43000d7a949ff1bd8742fcad57434d08b2455bfc5f681dd0cf0a32f6",
		"0405269acdc54c24220f67911a3ac709b129ff2454717875b789288954bb2e6afaed4d59ff0f4c4d14afff9f6f4e1ffb71a1e5457fa2ca7440a020218559ab7f3f"}
)

func GetNode(arr []string, index int) (string, error) {
	for k, v := range arr {
		if k == index {
			return v, nil
		}
	}
	return "", errors.New(fmt.Sprintf("GetIndex arr:%v p:%v", arr, index))
}

func byte2Int(data []byte) uint32 {
	var buffer bytes.Buffer
	_, err := buffer.Write(data)
	if err != nil {
		return 0
	}
	var i uint16
	err = binary.Read(&buffer, binary.BigEndian, &i)
	if err != nil {
		return 0
	}
	return uint32(i)
}
