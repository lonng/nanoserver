package provider

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lonng/nanoserver/db/model"
	"github.com/lonng/nanoserver/pkg/algoutil"
	"github.com/lonng/nanoserver/pkg/errutil"
	"github.com/lonng/nanoserver/protocol"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type wechat struct {
	appkey        string
	appId         string
	merId         string
	unifyOrderURL string
	callbackURL   string
}

var Wechat = &wechat{}

// 请求https://api.mch.weixin.qq.com/pay/unifiedorder需要填入的参数
// refs: https://pay.weixin.qq.com/wiki/doc/api/app/app.php?chapter=9_1
type UnifyOrderReq struct {
	Appid          string `xml:"appid"`
	Body           string `xml:"body"`
	MchID          string `xml:"mch_id"`
	NonceStr       string `xml:"nonce_str"`
	NotifyURL      string `xml:"notify_url"`
	TradeType      string `xml:"trade_type"`
	SpbillCreateIP string `xml:"spbill_create_ip"`
	TotalFee       int    `xml:"total_fee"`
	OutTradeNo     string `xml:"out_trade_no"`
	Sign           string `xml:"sign"`
}

type UnifyOrderResp struct {
	Return_code string `xml:"return_code"`
	Return_msg  string `xml:"return_msg"`
	Appid       string `xml:"appid"`
	MchID       string `xml:"mch_id"`
	NonceStr    string `xml:"nonce_str"`
	Sign        string `xml:"sign"`
	ResultCode  string `xml:"result_code"`
	PrepayID    string `xml:"prepay_id"`
	TradeType   string `xml:"trade_type"`
	ErrCode     string `xml:"err_code"`
}

// 微信支付计算签名的函数
func signCalculator(params map[string]interface{}, key string) (string, error) {
	if params == nil || key == "" {
		return "", errutil.ErrIllegalParameter
	}

	sortedKeys := make([]string, 0)
	for k := range params {
		sortedKeys = append(sortedKeys, k)
	}

	sort.Strings(sortedKeys)

	buf := &bytes.Buffer{}
	for _, k := range sortedKeys {
		if params[k] == nil {
			continue
		}

		switch params[k].(type) {
		case string:
			if len(params[k].(string)) == 0 {
				continue
			}

		case int:
			if params[k].(int) == 0 {
				continue
			}
		}

		fmt.Fprintf(buf, "%s=%v&", k, params[k])
	}
	fmt.Fprintf(buf, "key=%s", key)
	md5Ctx := md5.New()
	signStr := buf.Bytes()
	md5Ctx.Write(signStr)
	cipherStr := md5Ctx.Sum(nil)
	sign := strings.ToUpper(hex.EncodeToString(cipherStr))
	return sign, nil
}

// 微信支付签名验证函数
func verify(m map[string]interface{}, key, sign string) bool {
	signed, _ := signCalculator(m, key)
	return sign == signed
}

func (wc *wechat) CreateOrderResponse(order *model.Order) (interface{}, error) {
	req := UnifyOrderReq{
		Appid:          wc.appId,             //微信开放平台的app的appid
		Body:           order.ProductName,    //产品名
		MchID:          wc.merId,             //商户ID
		NonceStr:       algoutil.RandStr(32), //随机数
		NotifyURL:      wc.callbackURL,
		TradeType:      "APP",
		SpbillCreateIP: strings.Split(order.Ip, ":")[0],
		TotalFee:       order.ProductCount * 1,
		OutTradeNo:     order.OrderId,
	}

	m := make(map[string]interface{}, 0)

	m["appid"] = req.Appid
	m["body"] = req.Body
	m["mch_id"] = req.MchID
	m["notify_url"] = req.NotifyURL
	m["trade_type"] = req.TradeType
	m["spbill_create_ip"] = req.SpbillCreateIP
	m["total_fee"] = req.TotalFee
	m["out_trade_no"] = req.OutTradeNo
	m["nonce_str"] = req.NonceStr

	sign, err := signCalculator(m, wc.appkey)
	if err != nil {
		return nil, err
	}

	req.Sign = sign
	bytesReq, err := xml.Marshal(req)
	if err != nil {
		return nil, err
	}
	//unified order接口需要http body中xmldoc的根节点是<xml></xml>这种，所以这里需要replace一下
	strReq := strings.Replace(string(bytesReq), "UnifyOrderReq", "xml", -1)
	bytesReq = []byte(strReq)

	log.Debugf("prepay id request: %s", strReq)

	request, err := http.NewRequest("POST", wc.unifyOrderURL, bytes.NewReader(bytesReq))
	if err != nil {
		log.Errorf("create unify order failed: %s", err.Error())
		return nil, err
	}
	request.Header.Set("Accept", "application/xml")
	request.Header.Set("Content-Type", "application/xml;charset=utf-8")
	c := http.Client{}
	response, err := c.Do(request)
	if err != nil {
		log.Errorf("request unify order failed: %s", err.Error())
		return nil, err
	}

	println(response.StatusCode)

	xmlResp := &UnifyOrderResp{}
	if err := xml.NewDecoder(response.Body).Decode(xmlResp); err != nil {
		log.Errorf("unify order response unmarsharl failed: %s", err.Error())
		return nil, err
	}

	const (
		fail = "FAIL"
	)

	log.Debugf("prepay id response: %+v", xmlResp)

	if xmlResp.Return_code == fail {
		fmt.Println("unify order request prepay id failed: " + xmlResp.Return_msg)
		return nil, errutil.ErrRequestPrePayIDFailed
	}

	m = make(map[string]interface{}, 0)

	ret := protocol.CreateOrderWechatReponse{
		OrderId:   order.OrderId,
		PrePayID:  xmlResp.PrepayID,
		AppID:     xmlResp.Appid,
		PartnerId: xmlResp.MchID,
		//Package: "Sign=WXPay",
		NonceStr:  algoutil.RandStr(32),
		Timestamp: strconv.FormatInt(time.Now().Unix(), 10),
		Extra:     order.Extra,
	}

	m["appid"] = ret.AppID
	m["partnerid"] = ret.PartnerId
	m["prepayid"] = ret.PrePayID
	m["noncestr"] = ret.NonceStr
	m["timestamp"] = ret.Timestamp
	m["package"] = "Sign=WXPay"

	ret.Sign, err = signCalculator(m, wc.appkey)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

const format = "20060102150405"

func (wc *wechat) Notify(request *protocol.WechatOrderCallbackRequest) (*model.Trade, interface{}, error) {
	var reqMap map[string]interface{}
	reqMap = make(map[string]interface{}, 0)

	reqMap["return_code"] = request.ReturnCode
	reqMap["return_msg"] = request.ReturnMsg
	reqMap["appid"] = request.Appid
	reqMap["mch_id"] = request.MchID
	reqMap["device_info"] = request.DeviceInfo
	reqMap["nonce_str"] = request.Nonce
	//sign
	reqMap["result_code"] = request.ResultCode
	reqMap["err_code"] = request.ErrCode
	reqMap["err_code_des"] = request.ErrCodeDes

	reqMap["openid"] = request.Openid
	reqMap["is_subscribe"] = request.IsSubscribe
	reqMap["trade_type"] = request.TradeType
	reqMap["bank_type"] = request.BankType
	reqMap["total_fee"] = request.TotalFee
	reqMap["fee_type"] = request.FeeType
	reqMap["cash_fee"] = request.CashFee
	reqMap["cash_fee_type"] = request.CashFeeType

	reqMap["coupon_fee"] = request.CouponFee
	reqMap["coupon_count"] = request.CouponCount
	reqMap["coupon_id_$n"] = request.CouponIDDollarN
	reqMap["coupon_fee_$n"] = request.CouponFeeDollarN

	reqMap["transaction_id"] = request.TransactionID
	reqMap["out_trade_no"] = request.OutTradeNo
	reqMap["attach"] = request.Attach
	reqMap["time_end"] = request.TimeEnd

	var resp protocol.WechatOrderCallbackResponse
	if verify(reqMap, wc.appkey, request.Sign) {
		resp.ReturnCode = "SUCCESS"
		resp.ReturnMsg = "OK"
	} else {
		resp.ReturnCode = "FAIL"
		resp.ReturnMsg = "failed to verify sign, please retry!"
	}

	//bytes, err := xml.Marshal(resp)
	//cbResp := strings.Replace(string(bytes), "WechatOrderCallbackResponse", "xml", -1)
	//if err != nil {
	//	return nil, nil, err
	//}

	format := "<xml><return_code><![CDATA[%s]]></return_code><return_msg><![CDATA[%s]]></return_msg></xml>"

	cbResp := fmt.Sprintf(format, resp.ReturnCode, resp.ReturnMsg)

	fmt.Println("response: ", cbResp)
	trade := &model.Trade{}
	trade.PayOrderId = request.TransactionID
	trade.OrderId = request.OutTradeNo

	createAt, err := time.Parse(format, request.TimeEnd)
	if err != nil {
		createAt = time.Now()
	}
	trade.PayCreateAt = createAt.Unix()

	payAt, err := time.Parse(format, request.TimeEnd)
	if err != nil {
		payAt = time.Now()
	}
	trade.PayAt = payAt.Unix()

	trade.MerchantId = request.MchID
	trade.ComsumerId = request.Openid
	trade.Raw = request.Raw
	return trade, cbResp, nil
}

func (wc *wechat) Setup() error {
	log.Info("pay_provider: wechat setup")

	var (
		appId         = viper.GetString("wechat.appid")
		appKey        = viper.GetString("wechat.appsecret")
		merId         = viper.GetString("wechat.mer_id")
		unifyOrderURL = viper.GetString("wechat.unify_order_url")
		callbackURL   = viper.GetString("wechat.callback_url")
	)
	if unifyOrderURL == "" || callbackURL == "" || appId == "" || appKey == "" || merId == "" {
		log.Debugf("appId=%s, appKey=%s, merId=%s, unifyOrderURL=%s, callbackURL=%s", appId,
			appKey,
			merId,
			unifyOrderURL,
			callbackURL)
		return errors.New("the wechat's config is invalid")
	}
	wc.appId = appId
	wc.appkey = appKey
	wc.merId = merId
	wc.callbackURL = callbackURL
	wc.unifyOrderURL = unifyOrderURL
	return nil
}
