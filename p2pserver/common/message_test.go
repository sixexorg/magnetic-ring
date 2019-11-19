/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */
//  github.com/stretchr/testify/assert

package common

import (
	"bytes"
	"net"
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
	comm "github.com/sixexorg/magnetic-ring/p2pserver/common"
	// cm "github.com/ontio/ontology/common"
)

func TestVerackSerializationDeserialization(t *testing.T) {
	var msg VerACK
	msg.IsConsensus = false

	MessageTest(t, &msg)
}

func TestAddrReqSerializationDeserialization(t *testing.T) {
	var msg AddrReq

	MessageTest(t, &msg)
}

// addr
func MessageTest(t *testing.T, msg Message) {
	sink := common.NewZeroCopySink(nil)
	err := WriteMessage(sink, msg)
	assert.Nil(t, err)

	demsg, _, err := ReadMessage(bytes.NewBuffer(sink.Bytes()))
	assert.Nil(t, err)

	assert.Equal(t, msg, demsg)
}

func TestAddressSerializationDeserialization(t *testing.T) {
	var msg Addr
	var addr [16]byte
	ip := net.ParseIP("192.168.0.1")
	ip.To16()
	copy(addr[:], ip[:16])
	nodeAddr := comm.PeerAddr{
		Time:          12345678,
		Services:      100,
		IpAddr:        addr,
		Port:          8080,
		ConsensusPort: 8081,
		ID:            987654321,
	}
	msg.NodeAddrs = append(msg.NodeAddrs, nodeAddr)

	MessageTest(t, &msg)
}

// ping
func TestPingSerializationDeserialization(t *testing.T) {
	var msg Ping
	msg.Height = 1

	MessageTest(t, &msg)
}

//pong
func TestPongSerializationDeserialization(t *testing.T) {
	var msg Pong
	msg.Height = 1

	MessageTest(t, &msg)
}

//getheaders
func TestBlkHdrReqSerializationDeserialization(t *testing.T) {
	var msg HeadersReq
	msg.Len = 1

	hashstr := "8932da73f52b1e22f30c609988ed1f693b6144f74fed9a2a20869afa7abfdf5e"
	msg.HashStart, _ = common.Uint256FromHexString(hashstr)

	MessageTest(t, &msg)
}

//headers

// inv

// getdata
func TestBlkReqSerializationDeserialization(t *testing.T) {
	var msg BlocksReq
	msg.HeaderHashCount = 1

	hashstr := "8932da73f52b1e22f30c609988ed1f693b6144f74fed9a2a20869afa7abfdf5e"
	msg.HashStart, _ = common.Uint256FromHexString(hashstr)

	MessageTest(t, &msg)
}

func TestDataReqSerializationDeserialization(t *testing.T) {
	var msg DataReq
	msg.DataType = 0x02

	hashstr := "8932da73f52b1e22f30c609988ed1f693b6144f74fed9a2a20869afa7abfdf5e"
	bhash, _ := common.HexToBytes(hashstr)
	copy(msg.Hash[:], bhash)

	MessageTest(t, &msg)
}

// notfound
func Uint256ParseFromBytes(f []byte) common.Uint256 {
	if len(f) != 32 {
		return common.Uint256{}
	}

	var hash [32]uint8
	for i := 0; i < 32; i++ {
		hash[i] = f[i]
	}
	return common.Uint256(hash)
}

func TestNotFoundSerializationDeserialization(t *testing.T) {
	var msg NotFound
	str := "123456"
	hash := []byte(str)
	msg.Hash = Uint256ParseFromBytes(hash)
	MessageTest(t, &msg)
}
