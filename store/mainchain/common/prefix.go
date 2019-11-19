package common

type PrefixKey byte

const (
	DATA_BLOCK       PrefixKey = 0x00
	DATA_HEADER                = 0x01
	DATA_TRANSACTION           = 0x02
	DATA_SIGDATA               = 0x03

	IX_HEADER_HASH_LIST PrefixKey = 0x09 //Block height => block hash key prefix

	//SYSTEM
	SYS_CURRENT_BLOCK      PrefixKey = 0x10 //Current block key prefix
	SYS_VERSION            PrefixKey = 0x11 //Store version key prefix
	SYS_CURRENT_STATE_ROOT PrefixKey = 0x12 //no use
	SYS_BLOCK_MERKLE_TREE  PrefixKey = 0x13 // Block merkle tree root key prefix

	LEAGUE_BIRTH_CERT PrefixKey = 0x40

	ST_ACCOUNT     PrefixKey = 0x50
	ST_LEAGUE      PrefixKey = 0x51
	ST_MEMBER      PrefixKey = 0x52
	ST_RECEIPT     PrefixKey = 0x53
	ST_LEAGUE_META PrefixKey = 0X54
	ST_SYMBOL      PrefixKey = 0X55

	//external league
	EXT_LEAGUE_DATA_BLOCK    PrefixKey = 0x70
	EXT_LEAGUE_DATA_HEADER   PrefixKey = 0x71
	EXT_LEAGUE_ACCOUNT       PrefixKey = 0x72
	EXT_LEAGUE_MAIN_USED     PrefixKey = 0x73
	EXT_LEAGUE_ACCOUNT_INDEX PrefixKey = 0X74
	EXT_VOTE_STATE           PrefixKey = 0X75
	EXT_VOTE_RECORD          PrefixKey = 0X76
	EXT_VOTE_TX              PrefixKey = 0X77
	EXT_LEAGUE_FULL_TX       PrefixKey = 0x78
	EXT_UT                   PrefixKey = 0x79

	Accts_Lvl_All PrefixKey = 0x90 //No need to grow
	Accts_Lvl_Act PrefixKey = 0x91 //Long growing and can continue to grow
	Accts_Lvl_Map PrefixKey = 0x92
	ACCT_LVL      PrefixKey = 0X93

	ACCOUNT_HASHES PrefixKey = 0xa1
)
