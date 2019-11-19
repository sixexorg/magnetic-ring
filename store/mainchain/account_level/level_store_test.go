package account_level

import (
	"os"
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/mock"
)

func TestLevelStore(t *testing.T) {
	dbDir := "./test"
	os.RemoveAll(dbDir)
	s, _ := NewLevelStore(dbDir)
	m1 := make(map[common.Address]EasyLevel)
	m1[mock.Address_1] = lv5
	s.SaveLevels(50, m1)
	m2 := make(map[common.Address]EasyLevel)
	m2[mock.Address_1] = lv7
	s.SaveLevels(70, m2)
	m3 := make(map[common.Address]EasyLevel)
	m3[mock.Address_1] = lv8
	s.SaveLevels(75, m3)
	m4 := make(map[common.Address]EasyLevel)
	m4[mock.Address_1] = lv9
	s.SaveLevels(80, m4)
	assert.Equal(t, s.GetAccountLevel(5, mock.Address_1), lv0)
	assert.Equal(t, s.GetAccountLevel(50, mock.Address_1), lv5)
	assert.Equal(t, s.GetAccountLevel(55, mock.Address_1), lv5)
	assert.Equal(t, s.GetAccountLevel(73, mock.Address_1), lv7)
	assert.Equal(t, s.GetAccountLevel(76, mock.Address_1), lv8)
	assert.Equal(t, s.GetAccountLevel(89, mock.Address_1), lv9)
	ftress1 := []common.FTreer{
		&HeightLevel{Height: 50, Lv: lv5},
		&HeightLevel{Height: 70, Lv: lv7},
	}

	ftrees := s.GetAccountLvlRange(15, 74, mock.Address_1)
	assert.Equal(t, ftress1, ftrees)
}
