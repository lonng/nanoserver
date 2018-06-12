package algoutil

import (
	"testing"
)

type Test struct {
	Foo, Bar, FooBar string
}

func TestParamsToStruct(t *testing.T) {
	test := &Test{}
	ParamsToStruct("foo=hello&bar=world&foobar=helloworld", test)
	if test.Foo != "hello" || test.Bar != "world" || test.FooBar != "helloworld" {
		t.Fail()
	}
}

func TestSortParams(t *testing.T) {
	if SortParams(ParseParams("zdf=sdff&b=c&a=c&aaaaa=cdf")) != "a=c&aaaaa=cdf&b=c&zdf=sdff" {
		t.Fail()
	}
	unsorted := "discount=0.00&payment_type=1&subject=金币&trade_no=2016031421001004170242826341&buyer_email=15520707860&gmt_create=2016-03-14 16:10:24&notify_type=trade_status_sync&quantity=1&out_trade_no=426006374921535488&seller_id=2088801054724902&notify_time=2016-03-14 16:10:25&body=500金币&trade_status=TRADE_SUCCESS&is_total_fee_adjust=N&total_fee=0.01&gmt_payment=2016-03-14 16:10:25&seller_email=854761339@qq.com&price=0.01&buyer_id=2088402810244179&notify_id=cfdcab45ae04498a568e03357edd6cehba&use_coupon=N&sign_type=RSA&sign=e+VYyBpGyLubIdOMpQt0B/StveBZgthVcTbsNEFFmUgjJ+Bahl+3pf5g0Dim1yBZ3Sxz4C57qrozqfzjdVhWf/SEs8QWyf6+V4LgcosTTWpXBLXnVWfMvInGuxOUrufh9fG874tIISMkPPrkud+vsTn6wcqetipBh+wM+P7J9NI="
	sorted := "body=500金币&buyer_email=15520707860&buyer_id=2088402810244179&discount=0.00&gmt_create=2016-03-14 16:10:24&gmt_payment=2016-03-14 16:10:25&is_total_fee_adjust=N&notify_id=cfdcab45ae04498a568e03357edd6cehba&notify_time=2016-03-14 16:10:25&notify_type=trade_status_sync&out_trade_no=426006374921535488&payment_type=1&price=0.01&quantity=1&seller_email=854761339@qq.com&seller_id=2088801054724902&sign=e+VYyBpGyLubIdOMpQt0B/StveBZgthVcTbsNEFFmUgjJ+Bahl+3pf5g0Dim1yBZ3Sxz4C57qrozqfzjdVhWf/SEs8QWyf6+V4LgcosTTWpXBLXnVWfMvInGuxOUrufh9fG874tIISMkPPrkud+vsTn6wcqetipBh+wM+P7J9NI=&sign_type=RSA&subject=金币&total_fee=0.01&trade_no=2016031421001004170242826341&trade_status=TRADE_SUCCESS&use_coupon=N"
	if SortParams(ParseParams(unsorted)) != sorted {
		t.Fail()
	}
}
