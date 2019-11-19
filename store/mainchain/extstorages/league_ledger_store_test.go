package extstorages

import (
	"testing"

	"github.com/sixexorg/magnetic-ring/mock"
)

func TestGetExtDataByHeight(t *testing.T) {
	extData, err := leadger.GetExtDataByHeight(mock.LeagueAddress1, 1)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	t.Log(extData)
}
