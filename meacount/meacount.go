package meacount

import "github.com/sixexorg/magnetic-ring/account"

var owner account.Account

func SetOwner(own account.Account) {
	owner = own
}

func GetOwner() account.Account {
	return owner
}
