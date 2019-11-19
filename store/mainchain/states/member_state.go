package states

import (
	"io"

	"io/ioutil"

	"sort"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/errors"
)

//for leagueState memberRoot
type LeagueMember struct {
	LeagueId common.Address
	Height   uint64
	hash     common.Hash
	Data     *LeagueAccount
}

type LeagueAccount struct {
	Account common.Address
	Status  LeagueAccountStatus
}

type LeagueMembers []*LeagueMember

func (s LeagueMembers) Len() int           { return len(s) }
func (s LeagueMembers) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s LeagueMembers) Less(i, j int) bool { return s[i].hash.String() < s[j].hash.String() }
func (s LeagueMembers) GetHashRoot() common.Hash {
	sort.Sort(s)
	hashes := make([]common.Hash, 0, s.Len())
	for _, v := range s {
		hashes = append(hashes, v.Hash())
	}
	return common.ComputeMerkleRoot(hashes)
}

func (this *LeagueMember) Hash() common.Hash {
	buff := this.Serialization()
	hash, _ := common.ParseHashFromBytes(common.Sha256(buff))
	this.hash = hash
	return hash
}
func (this *LeagueMember) Serialization() []byte {
	sk := sink.NewZeroCopySink(nil)
	sk.WriteAddress(this.LeagueId)
	sk.WriteAddress(this.Data.Account)
	sk.WriteUint64(this.Height)
	sk.WriteByte(byte(this.Data.Status))
	return sk.Bytes()
}
func (this *LeagueMember) Deserialize(r io.Reader) error {
	buff, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	var (
		eof    bool
		status byte
	)
	source := sink.NewZeroCopySource(buff)
	_, eof = source.NextByte()
	this.LeagueId, eof = source.NextAddress()
	this.Data = &LeagueAccount{}
	this.Data.Account, eof = source.NextAddress()
	this.Height, eof = source.NextUint64()

	status, eof = source.NextByte()
	if eof {
		return errors.ERR_TXRAW_EOF
	}
	this.Data.Status = LeagueAccountStatus(status)
	return nil
}
