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
package db

import (
	"fmt"
	"os"
	"testing"

	"bytes"

	"github.com/syndtr/goleveldb/leveldb/util"
)

var testLevelDB *LevelDBStore

func TestMain(m *testing.M) {
	dbFile := "./test"
	var err error
	testLevelDB, err = NewLevelDBStore(dbFile)
	if err != nil {
		fmt.Printf("NewLevelDBStore error:%s\n", err)
		return
	}
	m.Run()
	testLevelDB.Close()
	os.RemoveAll(dbFile)
	os.RemoveAll("ActorLog")
}

func TestLevelDB(t *testing.T) {
	key := "foo"
	value := "bar"
	err := testLevelDB.Put([]byte(key), []byte(value))
	if err != nil {
		t.Errorf("Put error:%s", err)
		return
	}
	v, err := testLevelDB.Get([]byte(key))
	if err != nil {
		t.Errorf("Get error:%s", err)
		return
	}
	if string(v) != value {
		t.Errorf("Get error %s != %s", v, value)
		return
	}
	err = testLevelDB.Delete([]byte(key))
	if err != nil {
		t.Errorf("Delete error:%s", err)
		return
	}
	ok, err := testLevelDB.Has([]byte(key))
	if err != nil {
		t.Errorf("Has error:%s", err)
		return
	}
	if ok {
		t.Errorf("Key:%s shoule delete", key)
		return
	}
}

func TestBatch(t *testing.T) {
	testLevelDB.NewBatch()

	key1 := "foo1"
	value1 := "bar1"
	testLevelDB.BatchPut([]byte(key1), []byte(value1))

	key2 := "foo2"
	value2 := "bar2"
	testLevelDB.BatchPut([]byte(key2), []byte(value2))

	err := testLevelDB.BatchCommit()
	if err != nil {
		t.Errorf("BatchCommit error:%s", err)
		return
	}

	v1, err := testLevelDB.Get([]byte(key1))
	if err != nil {
		t.Errorf("Get error:%s", err)
		return
	}
	if string(v1) != value1 {
		t.Errorf("Get %s != %s", v1, value1)
		return
	}
}

func TestIterator(t *testing.T) {
	kvMap := make(map[string]string)
	kvMap["foo0001"] = "bar0001"
	kvMap["foo0002"] = "bar0002"
	kvMap["foo0005"] = "bar0005"
	kvMap["foo0101"] = "bar0101"
	kvMap["foo1001"] = "bar1001"
	kvMap["foo0021"] = "bar0021"
	testLevelDB.NewBatch()
	for k, v := range kvMap {
		testLevelDB.BatchPut([]byte(k), []byte(v))
	}
	err := testLevelDB.BatchCommit()
	if err != nil {
		t.Errorf("BatchCommit error:%s", err)
		return
	}
	kvs := make(map[string]string)

	fmt.Println("--------------------iter1----------------------------")
	iter := testLevelDB.NewIterator([]byte("fo"))
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		kvs[string(key)] = string(value)
		fmt.Printf("Key:%s Value:%s\n", key, value)
	}
	iter.Release()

	fmt.Println("--------------------iter2----------------------------")
	iter2 := testLevelDB.NewSeniorIterator(&util.Range{Start: []byte("foo0002"), Limit: []byte("foo0010")})
	for iter2.Next() {
		key := iter2.Key()
		value := iter2.Value()
		fmt.Printf("Key:%s Value:%s\n", key, value)
	}
	/*	iter2.Prev()
		iter2.Prev()
		fmt.Printf("Key:%s Value:%s\n", iter2.Key(), iter2.Value())*/
	iter2.Release()

	fmt.Println("--------------------iter3----------------------------")
	iter3 := testLevelDB.NewSeniorIterator(util.BytesPrefix([]byte("foo0001")))
	for iter3.Next() {
		fmt.Printf("Key:%s Value:%s\n", iter3.Key(), iter3.Value())
	}

	iter3.Release()
	fmt.Println("--------------------iter4----------------------------")
	iter4 := testLevelDB.NewSeniorIterator(nil)
	/*	for ok := iter4.Seek([]byte("foo1001")); ok; ok = iter4.Prev() {
		fmt.Printf("查找数据:%s, value:%s\n", iter4.Key(), iter4.Value())
	}*/
	flag := true
	if iter4.Seek([]byte("foo0011")) {
		cmp := bytes.Compare(iter4.Key(), []byte("foo0003"))
		if cmp == 0 {
			fmt.Printf("查找数据1:%s, value:%s\n", iter4.Key(), iter4.Value())
			flag = false
		}
	}
	if flag && iter4.Prev() {
		fmt.Printf("查找数据2:%s, value:%s\n", iter4.Key(), iter4.Value())
	}
	iter4.Release()
}

func BenchmarkIterator(b *testing.B) {
	kvMap := make(map[string]string)
	kvMap["foo0001"] = "bar0001"
	kvMap["foo0002"] = "bar0002"
	kvMap["foo0005"] = "bar0005"
	kvMap["foo0101"] = "bar0101"
	kvMap["foo1001"] = "bar1001"
	kvMap["foo0021"] = "bar0021"
	testLevelDB.NewBatch()
	for k, v := range kvMap {
		testLevelDB.BatchPut([]byte(k), []byte(v))
	}
	err := testLevelDB.BatchCommit()
	if err != nil {
		b.Errorf("BatchCommit error:%s", err)
		return
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ { //use b.N for looping
		iter4 := testLevelDB.NewSeniorIterator(nil)
		/*	for ok := iter4.Seek([]byte("foo1001")); ok; ok = iter4.Prev() {
			fmt.Printf("Find data:%s, value:%s\n", iter4.Key(), iter4.Value())
		}*/
		flag := true
		if iter4.Seek([]byte("foo0011")) {
			cmp := bytes.Compare(iter4.Key(), []byte("foo0003"))
			if cmp == 0 {
				//fmt.Printf("Find data1:%s, value:%s\n", iter4.Key(), iter4.Value())
				flag = false
			}
		}
		if flag && iter4.Prev() {
			//fmt.Printf("Find data2:%s, value:%s\n", iter4.Key(), iter4.Value())
		}
		iter4.Release()

	}
}
