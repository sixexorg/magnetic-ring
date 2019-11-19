package common

type PrefixKey byte

const (
	DATA_BLOCK       PrefixKey = 0x00
	DATA_HEADER                = 0x01
	DATA_TRANSACTION           = 0x02

	IX_HEADER_HASH_LIST PrefixKey = 0x09 //Block height => block hash key prefix

	//SYSTEM
	SYS_CURRENT_BLOCK      PrefixKey = 0x10 //Current block key prefix
	SYS_VERSION            PrefixKey = 0x11 //Store version key prefix
	SYS_CURRENT_STATE_ROOT PrefixKey = 0x12 //no use
	SYS_BLOCK_MERKLE_TREE  PrefixKey = 0x13 // Block merkle tree root key prefix

	ST_ACCOUNT      PrefixKey = 0x50
	ST_ACCOUNT_ROOT PrefixKey = 0x51
	ST_VOTE_STATE   PrefixKey = 0x60
	ST_VOTE_RECORD  PrefixKey = 0x61
	ST_UT           PrefixKey = 0X70
)
