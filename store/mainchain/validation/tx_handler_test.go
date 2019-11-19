package validation

import (
	"testing"

	"github.com/sixexorg/magnetic-ring/mock"
)

func TestAnalysis_TansferBox(t *testing.T) {
	oplogs := AnalysisTansferBox(mock.Mock_Tx)
	for _, v := range oplogs {
		t.Log(v.method)
	}
}
