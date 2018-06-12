package db

import (
	"github.com/lonnng/nanoserver/internal/errutil"
	"github.com/lonnng/nanoserver/db/model"
)

//InsertApp add a new app
func InsertApp(app *model.App) error {
	if app == nil {
		return nil
	}
	_, err := DB.Insert(app)
	if err != nil {
		err = errutil.YXErrDBOperation
	}
	return err
}

//QueryApp get the app by app id
func QueryApp(appid string) (*model.App, error) {
	app := &model.App{}
	has, err := DB.Where("appid=?", appid).Get(app)
	if !has {
		err = errutil.YXErrNotFound
	}
	if err != nil {
		err = errutil.YXErrDBOperation
		return nil, err
	}
	if app.Status != StatusNormal {
		return nil, err
	}
	return app, err
}

//DeleteApp delete the specific app
func DeleteApp(appid string) error {
	_, err := DB.Exec(`
		UPDATE app
		SET status = ?
		WHERE appid = ?`, StatusDeleted, appid)
	if err != nil {
		err = errutil.YXErrDBOperation
	}
	return err
}

//UpdateApp update the app's attribute(s)
func UpdateApp(app *model.App) error {
	if app == nil {
		return nil
	}
	_, err := DB.Where("appid=?", app.Appid).Update(app)
	if err != nil {
		err = errutil.YXErrDBOperation
	}
	return err
}

func QueryAppList(offset, count int, cpID string) ([]model.App, int64, error) {
	app := &model.App{
		Status: StatusNormal,
		CpId:   cpID,
	}

	total, err := DB.Count(app)
	if err != nil {
		return nil, 0, errutil.YXErrDBOperation
	}

	// retrieve all record
	if count == -1 {
		count = int(total)
	}

	result := make([]model.App, 0)
	err = DB.Where("status=?", StatusNormal).Desc("id").Limit(count, offset).Find(&result, app)

	if err != nil {
		return nil, 0, errutil.YXErrDBOperation
	}
	return result, total, nil
}

//AppKeyPairs return all appkey&model.Appsecret pairs
func AppKeyPairs() (map[string]*KeyPair, error) {
	var apps []model.App
	err := DB.Cols("appid", "app_key", "app_secret").Where("status=?", 1).Find(&apps)
	if err != nil {
		logger.Error( err)
		return nil, errutil.YXErrDBOperation
	}
	ret := make(map[string]*KeyPair)
	for _, app := range apps {
		ret[app.Appid] = &KeyPair{
			PublicKey:  string(app.AppKey),
			PrivateKey: string(app.AppSecret),
		}
	}
	return ret, nil
}

//KeyPairForApp return appkey&model.Appsecret pair for the app
func KeyPairForApp(appID string) (*KeyPair, error) {
	app := model.App{}
	has, err := DB.Cols("app_key", "app_secret").Where("appid=?", appID).Get(&app)
	if !has {
		return nil, errutil.YXErrDBOperation
	}

	if err != nil {
		return nil, err
	}
	kp := &KeyPair{
		PublicKey:  app.AppKey,
		PrivateKey: app.AppSecret,
	}
	return kp, nil
}
