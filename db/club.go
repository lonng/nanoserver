package db

import (
	"errors"
	"fmt"
	"time"

	"github.com/lonng/nanoserver/db/model"
)

func IsClubMember(clubId, uid int64) bool {
	uc := model.UserClub{
		Uid:    uid,
		ClubId: clubId,
		Status: model.UserClubStatusAgree,
	}

	has, err := database.Get(&uc)
	if err != nil {
		return false
	}
	return has
}

func IsBalanceEnough(clubId int64) bool {
	c := model.Club{ClubId: clubId}
	has, err := database.Get(&c)
	if err != nil {
		return false
	}
	if has == false {
		return false
	}
	return c.Balance > -100
}

func ApplyClub(uid, clubId int64) error {
	if clubId < 100000 || clubId >= 1000000 {
		return fmt.Errorf("俱乐部ID%d错误，请输入正确的俱乐部ID", clubId)
	}

	c := &model.Club{ClubId: clubId}
	ok, err := database.Get(c)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("ID为%d的俱乐部不存在，请检查是否输入错误", clubId)
	}

	uc := &model.UserClub{
		Uid:       uid,
		ClubId:    clubId,
		CreatedAt: time.Now().Unix(),
	}

	ok, err = database.Get(uc)
	if err != nil {
		return err
	}

	if ok {
		if uc.Status == model.UserClubStatusAgree {
			return errors.New("你已加入该俱乐部，无需申请")
		}

		if uc.Status == model.UserClubStatusApply {
			return errors.New("你已申请加入该俱乐部，等待部长同意")
		}
	}

	uc.Status = model.UserClubStatusApply
	_, err = database.Insert(uc)
	return err
}

func ClubList(uid int64) ([]model.Club, error) {
	bean := &model.UserClub{
		Uid:    uid,
		Status: model.UserClubStatusAgree,
	}

	list := []model.UserClub{}
	if err := database.Find(&list, bean); err != nil {
		return nil, err
	}

	if len(list) < 1 {
		return []model.Club{}, nil
	}

	ids := []int64{}
	for i := range list {
		ids = append(ids, list[i].ClubId)
	}

	ret := []model.Club{}
	database.In("club_id", ids).Find(&ret)

	return ret, nil
}

func ClubLoseBalance(clubId, balance int64, consume *model.CardConsume) error {
	session := database.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	c := &model.Club{ClubId: clubId}
	has, err := session.Get(c)
	if err != nil {
		return err
	}

	if !has {
		return fmt.Errorf("俱乐部不存在，ID=%d", clubId)
	}

	c.Balance -= balance

	//FIXED: 用户剩余1的时候, 扣除不成功
	if _, err := session.Cols("balance").Where("club_id=?", clubId).Update(c); err != nil {
		session.Rollback()
		return err
	}

	if _, err := session.Insert(consume); err != nil {
		session.Rollback()
		return err
	}

	return session.Commit()
}
