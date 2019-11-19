package errors

import "errors"

var (
	ERR_RADAR_ABNORMAL_DATA = errors.New("Context data does not match. Data exception.")

	ERR_RADAR_VERIFIED = errors.New("The radar is in calculation.")

	ERR_RADAR_MAINTX_BEFORE = errors.New("Tx has been processed.")

	ERR_RADAR_MAINTX_AFTER = errors.New("Before this tx, there's some tx that's missing.")

	ERR_RADAR_TX_MOT_MATCH = errors.New("Maintx does not match the leaguetX in it.")

	ERR_RADAR_TX_USELESS = errors.New("Useless tx")

	ERR_RADAR_M_HEIGHT_BLOCKING = errors.New("The local specific league validation height has not reached the specified height.")

	ERR_RADAR_NOT_FOUNT = errors.New("not found")
)
