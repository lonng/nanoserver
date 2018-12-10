package db

import (
	"github.com/lonng/nanoserver/db/model"
	"github.com/lonng/nanoserver/pkg/errutil"
)

func QueryThirdAccount(account, platform string) (*model.ThirdAccount, error) {
	t := &model.ThirdAccount{ThirdAccount: account, Platform: platform}
	has, err := database.Get(t)
	if err != nil {
		return nil, err
	}

	if !has {
		return nil, errutil.ErrThirdAccountNotFound
	}

	return t, nil
}

func InsertThirdAccount(account *model.ThirdAccount, u *model.User) error {
	session := database.NewSession()
	if err := session.Begin(); err != nil {
		return err
	}
	defer session.Close()

	if _, err := session.Insert(u); err != nil {
		session.Rollback()
		return err
	}

	// update uid
	account.Uid = u.Id

	if _, err := session.Insert(account); err != nil {
		session.Rollback()
		return err
	}

	return session.Commit()
}

func UpdateThirdAccount(account *model.ThirdAccount) error {
	if account == nil {
		return errutil.ErrInvalidParameter
	}
	_, err := database.Where("id=?", account.Id).Update(account)
	return err
}
