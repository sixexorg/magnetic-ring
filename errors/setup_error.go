package errors

import "errors"

var (
	ERR_SETUP_MAIN_LEDGER_UNINITIALIZED = errors.New("Ledger in the main chain was not initialized,so call to get the singleton failed")
)
