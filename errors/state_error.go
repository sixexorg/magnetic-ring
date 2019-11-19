package errors

import (
	"errors"
)

var (
	ERR_STATE_ACCOUNT_EMPTY = errors.New("Data does not exist.")

	ERR_STATE_DIRTY_DATA = errors.New("No such data is defined.")
	ERR_STATE_LACK_GAS   = errors.New("Insufficient account Energy.")
	ERR_STATE_GAS_OVER   = errors.New("Energy not enough.")
	ERR_STATE_LACK_BOX   = errors.New("Insufficient account crystal.")

	ERR_STATE_NONCE_USED                  = errors.New("Account state nonce has been used.")
	ERR_STATE_ACCOUNT_NONCE_DISCONTINUITY = errors.New("account state nonce discontinuity.")
	ERR_STATE_BONUS_NOT_ENOUGH            = errors.New("The bonus was not enough to cover the transaction costs.")
	ERR_STATE_LEAGUE_NONCE_BIGGER         = errors.New("league state nonce more bigger.")
	ERR_STATE_LEAGUE_NONCE_DISCONTINUITY  = errors.New("league state nonce discontinuity.")

	ERR_STATE_LEAGUE_RATE_DIFF = errors.New("league rate not match.")

	ERR_STATE_LEAGUE_ONCE = errors.New("an account can only perform one operation about add or exit in league.")

	ERR_STATE_SYMBOL_EXISTS = errors.New("symbol already exists.")

	ERR_STATE_LEAGUE_EXISTS = errors.New("league already exists.")
	ERR_STATE_FEE_NOT_VAILD = errors.New("Fee not vaild.")

	ERR_STATE_MEMBER_ILLEGAL_STATUS = errors.New("Specifies that the state is illegal. Use another state.")
)
