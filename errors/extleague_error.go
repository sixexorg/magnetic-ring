package errors

import "errors"

var (
	ERR_EXTERNAL_HEIGHT_NOT_MATCH = errors.New("this blockheight is not continuous with the current.")

	ERR_EXTERNAL_PREVBLOCKHASH_DIFF = errors.New("prevblockhash doesn't match")
	ERR_EXTERNAL_LEAGUE_DIFF        = errors.New("League transaction analysis failed.")

	ERR_EXTERNAL_TXROOT_DIFF = errors.New("The transactions in the block doesn't match the hashroot in the header.")

	ERR_EXTERNAL_BLOCK_DISGUISE = errors.New("This block is in disguise.")

	ERR_EXTERNAL_TX_REFERENCE_WRONG = errors.New("The transaction referenced is wrong.")

	ERR_EXTERNAL_MAIN_TX_USED = errors.New("The tx has been used.")
)
