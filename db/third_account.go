package db

import (
	"github.com/lonnng/nanoserver/internal/errutil"
	"github.com/lonnng/nanoserver/db/model"
)

func QueryThirdAccount(account, platform string) (*model.ThirdAccount, error) {
	t := &model.ThirdAccount{ThirdAccount: account, Platform: platform}
	has, err := DB.Get(t)
	if err != nil {
		return nil, err
	}

	if !has {
		return nil, errutil.YXErrThirdAccountNotFound
	}

	return t, nil
}

func InsertThirdAccount(account *model.ThirdAccount, u *model.User) error {
	session := DB.NewSession()
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
		return errutil.YXErrInvalidParameter
	}
	_, err := DB.Where("id=?", account.Id).Update(account)
	return err
}
