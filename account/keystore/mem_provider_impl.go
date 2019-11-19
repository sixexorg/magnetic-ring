// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//

package account

import (
	"errors"
	"fmt"

	"sync"

	"github.com/sixexorg/magnetic-ring/crypto"
	"github.com/sixexorg/magnetic-ring/crypto/cipher"
)

var (
	// ErrNeedAlias need alias
	ErrNeedAlias = errors.New("need alias")

	// ErrNotFound not find key
	ErrNotFound = errors.New("key not found")
)

//在内存中保存的私钥
// Entry keeps in memory
type Entry struct {
	key  crypto.PrivateKey
	data []byte
}

// MemoryProvider handle keystore with ecdsa
type MemoryProvider struct {

	// name of ecdsa provider
	name string

	// version of ecdsa provider
	version float32

	// a map storage entry
	entries map[string]Entry

	// encrypt key
	cipher *cipher.Cipher

	mu sync.RWMutex
}

// NewMemoryProvider generate a provider with version
func NewMemoryProvider(v float32) *MemoryProvider {
	p := &MemoryProvider{name: "memoryProvider", version: v, entries: make(map[string]Entry)}
	p.cipher = cipher.NewCipher()
	return p
}

// Aliases all entry in provider save
func (p *MemoryProvider) Aliases() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	aliases := []string{}
	for a := range p.entries {
		aliases = append(aliases, a)
	}
	return aliases
}

// SetKey assigns the given key (that has already been protected) to the given alias.
func (p *MemoryProvider) SetKey(a string, key crypto.PrivateKey, passphrase []byte) error {
	if len(a) == 0 {
		return ErrNeedAlias
	}
	if len(passphrase) == 0 {
		return errors.New("password length must > 0")
	}

	encoded := key.Bytes()


	data, err := p.cipher.Encrypt(encoded, passphrase)
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	entry := Entry{key, data}
	p.entries[a] = entry
	return nil
}

// GetKey returns the key associated with the given alias, using the given
// password to recover it.
func (p *MemoryProvider) GetKey(a string, passphrase []byte) (crypto.PrivateKey, error) {
	if len(a) == 0 {
		return nil, ErrNeedAlias
	}
	if len(passphrase) == 0 {
		return nil, errors.New("need password")
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	entry, ok := p.entries[a]
	if !ok {
		fmt.Printf("final here 001\n")
		return nil, ErrNotFound
	}
	//解密出私钥
	data, err := p.cipher.Decrypt(entry.data, passphrase)
	if err != nil {
		return nil, err
	}


	entry.key,err = crypto.ToPrivateKey(data)

	if err != nil {
		return nil, err
	}

	return entry.key, nil
}

// Delete remove key
func (p *MemoryProvider) Delete(a string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if &a == nil {
		return ErrNeedAlias
	}
	delete(p.entries, a)
	return nil
}

// ContainsAlias check provider contains key
func (p *MemoryProvider) ContainsAlias(a string) (bool, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if &a == nil {
		return false, ErrNeedAlias
	}

	if _, ok := p.entries[a]; ok {
		return true, nil
	}
	return false, ErrNotFound
}

// Clear clear all entries in provider
func (p *MemoryProvider) Clear() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.entries == nil {
		return errors.New("need entries map")
	}
	p.entries = make(map[string]Entry)
	return nil
}
