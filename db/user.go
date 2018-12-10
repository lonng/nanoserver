package db

import (
	"time"

	"fmt"
	"strconv"

	"github.com/lonng/nanoserver/db/model"
	"github.com/lonng/nanoserver/pkg/errutil"
	"github.com/lonng/nanoserver/protocol"
)

//QueryUser get the user by id
func QueryUser(id int64) (*model.User, error) {
	if id <= 0 {
		return nil, errutil.ErrUserNotFound
	}
	u := &model.User{
		Id: id,
	}
	has, err := database.Get(u)
	if !has {
		err = errutil.ErrUserNotFound
	}

	if err != nil {
		logger.Error(err)
		return nil, err
	}
	return u, nil

}

//UpdateUser update user's info
func UpdateUser(u *model.User) error {
	if u == nil {
		return nil
	}
	_, err := database.Where("id=?", u.Id).AllCols().Update(u)
	return err
}

//InsertUser insert a new user
func InsertUser(u *model.User) error {
	if u == nil {
		return nil
	}
	_, err := database.Insert(u)
	return err
}

//DeleteUser delete the user
func DeleteUser(uid int64) error {
	u := model.User{
		Status: StatusDeleted,
	}
	_, err := database.Where("uid=?", uid).Update(u)
	return err
}

func UserAddCoin(uid int64, coin int64) error {
	session := database.NewSession()
	defer session.Close()
	err := session.Begin()
	if err != nil {
		return errutil.ErrDBOperation
	}
	u := &model.User{Id: uid}
	has, err := session.Get(u)
	if err != nil {
		return err
	}
	if !has {
		return errutil.ErrNotFound
	}
	u.Coin += coin
	_, err = session.Where("uid=?", uid).Update(u)
	if err != nil {
		session.Rollback()
		return err
	}
	return session.Commit()
}

func UserLoseCoin(id int64, coin int64) error {
	user := &model.User{
		Id: id,
	}

	has, err := database.Get(user)
	if !has {
		return errutil.ErrNotFound
	}
	if err != nil {
		return err
	}

	return UserLoseCoinByUID(user.Id, coin)
}

func UserLoseCoinByUID(uid int64, coin int64) error {
	session := database.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		logger.Error(err.Error())
		return errutil.ErrDBOperation
	}
	u := &model.User{Id: uid}
	has, err := session.Get(u)
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	if !has {
		logger.Error("user not found")
		return errutil.ErrNotFound
	}
	if u.Coin < coin {
		logger.Error("coin not enough")
		return errutil.ErrCoinNotEnough
	}
	u.Coin -= coin
	//FIXED: 用户剩余1的时候, 扣除不成功
	if _, err := session.Cols("coin").Where("id=?", uid).Update(u); err != nil {
		session.Rollback()
		logger.Error(err.Error())
		return err
	}

	return session.Commit()
}

func InsertRegister(reg *model.Register) {
	chWrite <- reg
}

func userOnline(uid int64) error {
	u := &model.User{IsOnline: UserOnline, LastLoginAt: time.Now().Unix()}
	if _, err := database.Where("id=?", uid).Update(u); err != nil {
		return err
	}
	return nil
}

func QueryGuestUser(appId string, imei string) (*model.User, error) {
	bean := &model.Register{
		Imei:  imei,
		AppId: appId,
	}

	ok, err := database.Get(bean)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, errutil.ErrUserNotFound
	}

	user := &model.User{
		Id: bean.Uid,
	}
	ok, err = database.Get(user)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, errutil.ErrUserNotFound
	}

	return user, nil
}

func RegisterUserLog(u *model.User, d protocol.Device, appid string, channelID string, regType int) {
	// Insert user register record
	t := time.Now().Unix()
	reg := &model.Register{
		Uid:          u.Id,
		Remote:       d.Remote,
		Ip:           d.IP,
		Imei:         d.IMEI,
		Os:           d.OS,
		Model:        d.Model,
		AppId:        appid,
		ChannelId:    channelID,
		RegisterAt:   t,
		RegisterType: regType,
	}
	InsertRegister(reg)
}

func InsertLoginLog(uid int64, d protocol.Device, appid string, channelID string) {
	// Insert user operation record
	log := &model.Login{
		Uid:       uid,
		Remote:    d.Remote,
		Ip:        d.IP,
		Imei:      d.IMEI,
		Os:        d.OS,
		Model:     d.Model,
		AppId:     appid,
		ChannelId: channelID,
		LoginAt:   time.Now().Unix(),
	}
	userOnline(uid)
	chWrite <- log
}

//QueryUserInfo get the user by id
func QueryUserInfo(id int64) (*protocol.UserStatsInfo, error) {
	if id <= 0 {
		return nil, errutil.ErrUserNotFound
	}
	u := &model.User{
		Id: id,
	}
	has, err := database.Get(u)
	if !has {
		err = errutil.ErrUserNotFound
	}

	if err != nil {
		logger.Errorf("查询用户出错, Error%s", err.Error())
		return nil, err
	}

	//注册记录
	r := &model.Register{
		Uid: u.Id,
	}
	database.Get(r)

	//登录记录
	l := &model.Login{
		Uid: u.Id,
	}
	database.Desc("login_at").Get(l)

	ta := &model.ThirdAccount{
		Uid: u.Id,
	}
	database.Get(ta)

	//总对局数
	match, _ := database.Where("player0 =? OR player1 =? OR player2 = ?",
		id, id, id).Count(model.Desk{})

	usi := &protocol.UserStatsInfo{
		ID:             u.Id,
		Uid:            u.Id,
		Name:           ta.ThirdName,
		RegisterAt:     r.RegisterAt,
		RegisterIP:     r.Remote,
		LastestLoginAt: l.LoginAt,
		LastestLoginIP: l.Remote,
		RemainCard:     u.Coin,
		TotalMatch:     match,
		Stats:          make(map[int64]*protocol.DailyStats),
		StatsAt:        []int64{},
	}

	// 最多查最近一月的数据
	now := time.Now()
	temp := time.Date(now.Year(), now.Month()-1, now.Day(), 0, 0, 0, 0, time.Local)

	begin := temp.Unix()
	if begin < r.RegisterAt {
		t1 := time.Unix(r.RegisterAt, 0)
		begin = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, time.Local).Unix()
	}

	temp2 := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	f := func(id, begin, end int64) *protocol.DailyStats {
		ret := &protocol.DailyStats{}
		//桌数
		asCreator, _ := database.Where("creator = ? AND created_at BETWEEN ? AND ?",
			id,
			begin,
			end).Count(&model.Desk{})

		ret.AsCreator = asCreator

		//参与过的房号
		desks := []model.Desk{}
		database.Where("(player0 =? OR player1 =? OR player2 = ? ) AND created_at BETWEEN ? AND ?",
			id, id, id,
			begin, end,
		).Find(&desks)

		//胜局数
		wins := 0

		//战绩
		score := 0
		for _, d := range desks {

			if d.Player0 == id {
				if d.ScoreChange0 > 0 {
					wins++
				}
				score += d.ScoreChange0

			}
			if d.Player1 == id {
				if d.ScoreChange0 > 0 {
					wins++
				}
				score += d.ScoreChange1
			}
			if d.Player2 == id {
				if d.ScoreChange0 > 0 {
					wins++
				}
				score += d.ScoreChange2
			}

			ret.DeskNos = append(ret.DeskNos, d.DeskNo)

		}
		ret.Score = score
		ret.Win = wins

		return ret
	}

	for i := begin; i <= temp2.Unix(); i += 86400 {
		usi.Stats[i] = f(id, i, i+86400)
		usi.StatsAt = append(usi.StatsAt, i)
	}

	return usi, nil

}

//IsUserExists is the user with uid is existed
func IsUserExists(uid int64) bool {
	u := &model.User{
		Id: uid,
	}
	has, _ := database.Get(u)
	return has
}

func QueryUserList(offset, count int) ([]model.User, int64, error) {
	user := &model.User{
		Status: StatusNormal,
	}
	total, err := database.Count(user)
	if err != nil {
		logger.Error(err)
		return nil, 0, errutil.ErrDBOperation
	}
	result := make([]model.User, 0)
	err = database.Where("status=?", StatusNormal).Limit(count, offset).
		Desc("id").Find(&result)
	if err != nil {
		return nil, 0, errutil.ErrDBOperation
	}
	return result, total, nil
}

//注册用户数
func QueryRegisterUsers(begin, end int64) (int, error) {
	if begin > end {
		return 0, errutil.ErrIllegalParameter
	}

	user := &model.User{
		Status: StatusNormal,
	}

	total, err := database.Where("`register_at` BETWEEN ? AND ?", begin, end).Count(user)
	if err != nil {
		logger.Error(err)
		return 0, errutil.ErrDBOperation
	}

	return int(total), nil

}

//活跃人数
func QueryActivationUser(from, to int64) ([]*protocol.ActivationUser, error) {
	fn := func(from, to int64) *protocol.ActivationUser {
		mQuery, err := database.Query("SELECT COUNT(DISTINCT(uid)) AS users FROM `login` WHERE login_at BETWEEN ? AND ? ",
			from,
			to)

		cc := &protocol.ActivationUser{
			Date: from,
		}

		if len(mQuery) < 1 || err != nil {
			return cc
		}

		temp := string(mQuery[0]["users"])
		if temp != "" {
			cc.Value, err = strconv.ParseInt(temp, 10, 0)
		}

		return cc

	}
	begin := time.Unix(from, 0)

	var ret []*protocol.ActivationUser

	t := time.Date(begin.Year(), begin.Month(), begin.Day(), 0, 0, 0, 0, time.Local)

	for i := t.Unix(); i < to; i += dayInSecond {
		cc := fn(i, i+dayInSecond-1)

		ret = append(ret, cc)
	}
	return ret, nil
}

//实时在线人/桌数
func OnlineStatsLite() (*model.Online, error) {

	ol := &model.Online{}

	has, err := database.Desc("time").Get(ol)
	if err != nil || !has {
		return nil, err
	}

	return ol, err

}

func retentionHelper(current int) (*retentionStats, error) {
	f := func(step int) int64 {
		sql := fmt.Sprintf("SELECT COUNT( DISTINCT(login.uid)) AS retention FROM login JOIN register ON login.uid = register.uid	" +
			" WHERE register.register_at BETWEEN ? AND ? AND login.login_at BETWEEN ? AND ? ")

		fmt.Print(sql)
		fmt.Println(current, current+step)
		m, err := database.Query(
			sql,
			current,
			current+dayInSecond,
			current+step,
			current+step+dayInSecond)
		if len(m) < 1 || err != nil {
			return 0
		}

		retentionStr := string(m[0]["retention"])
		if retentionStr != "" {
			retention, err := strconv.ParseInt(retentionStr, 10, 0)
			if err != nil {
				return 0
			}

			return retention
		}

		return 0
	}

	r := &retentionStats{}

	r.loginDay1 = f(day1)
	r.loginDay2 = f(day2)
	r.loginDay3 = f(day3)
	r.loginDay7 = f(day7)
	r.loginDay14 = f(day14)
	r.loginDay30 = f(day30)

	sql := fmt.Sprintf("SELECT COUNT(DISTINCT(uid)) AS register FROM `register` WHERE register_at BETWEEN ? AND ? ")

	mQuery, err := database.Query(
		sql,
		current,
		current+dayInSecond,
	)

	if len(mQuery) < 1 || err != nil {
		return nil, err
	}

	registerStr := string(mQuery[0]["register"])
	if registerStr != "" {
		r.register, err = strconv.ParseInt(registerStr, 10, 0)
	}

	return r, nil
}

//某注册天的 n日留存
func RetentionList(current int) (*protocol.Retention, error) {
	st, err := retentionHelper(current)

	if err != nil {
		return nil, err
	}

	fill := func(rl *protocol.RetentionLite, register, login int64) {
		rl.Login = login

		if register != 0 {
			rl.Rate = fmt.Sprintf("%.2f", float32(login*100)/float32(register))
		} else {
			rl.Rate = "0.00"
		}

	}

	ret := &protocol.Retention{
		Date:     current,
		Register: st.register,
	}

	fill(&ret.Retention_1, st.register, st.loginDay1)
	fill(&ret.Retention_2, st.register, st.loginDay2)
	fill(&ret.Retention_3, st.register, st.loginDay3)
	fill(&ret.Retention_7, st.register, st.loginDay7)
	fill(&ret.Retention_14, st.register, st.loginDay14)
	fill(&ret.Retention_30, st.register, st.loginDay30)

	return ret, nil

}
