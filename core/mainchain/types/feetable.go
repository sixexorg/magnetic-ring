package types

const (
	Fee_Bottom       = uint64(100)
	Fee_Tansfer      = uint64(100)
	Fee_CreateLeague = uint64(10000)
)

var feeMap map[TransactionType]uint64

func GetFee(txTyp TransactionType) uint64 {
	return feeMap[txTyp]

}
func init() {
	feeMap = make(map[TransactionType]uint64, 10)
	feeMap[TransferBox] = Fee_Tansfer
	feeMap[TransferEnergy] = Fee_Tansfer
}
