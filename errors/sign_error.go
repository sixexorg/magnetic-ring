package errors

import "errors"

var (
	SIG_ERR_SIZE               = errors.New("sigpack size exception")
	SIG_ERR_APPR_LVL           = errors.New("approve level index out of bound")
	SIG_ERR_UNSIGN             = errors.New("sigbuf not found in sigdata")
	SIG_ERR_WRONGOWNER         = errors.New("not your transaction")
	SIG_ERR_CANOT              = errors.New("can not sign")
	SIG_ERR_NULL_UNITS         = errors.New("Multi Account Units can not be null")
	SIG_ERR_NOSIG              = errors.New("transaction has no signature")
	SIG_ERR_SIGS_NOTENOUGH     = errors.New("signature not enough")
	SIG_ERR_APPROVE_NOT_ENOUGH = errors.New("approve level sign not enough")
	SIG_ERR_INVALID_SIG        = errors.New("invalide signature")
	SIG_ERR_NOT_SUPPORT        = errors.New("not support this way")
)
