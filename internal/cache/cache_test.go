package cache

import (
	"os"
	"reflect"
	"testing"

	"github.com/lonnng/nanoserver/internal/algoutil"
	"github.com/lonnng/nanoserver/internal/types"

	"github.com/go-kit/kit/log"
)

func TestCRUD(t *testing.T) {
	var obj = struct {
		key string
		val int
	}{
		key: "object",
		val: 12315,
	}

	tables := map[string]interface{}{
		"":    "",
		"abc": "abc",
		"123": 123,
		"obj": obj,
	}

	for k, v := range tables {
		Delete(k)
		err := Set(k, v)
		if err != nil {
			t.Fatal(err)
		}

		_, err = Get(k)
		if err != nil {
			t.Fatal(err)
		}

		err = Delete(k)
		if err != nil {
			t.Fatal(err)
		}

		if Exists(k) {
			t.Fatalf("%s is existed", k)
		}

	}

}

func TestCURDStruct(t *testing.T) {
	tables := map[string]*types.UserMeta{
		"1": {1, 1, "123455"},
		"2": {2, 2, "123455"},
		"3": {3, 3, "123455"},
	}

	for k, v := range tables {
		Delete(k)

		err := SetStruct(k, v)
		if err != nil {
			t.Fatal(err)
		}

		meta := &types.UserMeta{}

		err = Struct(k, meta)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(meta, v) {
			t.Fatalf("value for %s is dismatch, want: %p got: %p\n", k, v, meta)
		}

		err = Delete(k)
		if err != nil {
			t.Fatal(err)
		}

		if Exists(k) {
			t.Fatalf("%s is existed", k)
		}
	}
}

func bootUpLogger() log.Logger {
	var logger log.Logger
	logger = log.NewLogfmtLogger(os.Stdout)
	logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC)
	logger = log.NewContext(logger).With("caller", log.Valuer(algoutil.CallSite))
	return logger

}

func TestMain(m *testing.M) {
	logger := bootUpLogger()
	closer := MustBootUp(logger, "master.1ktower.work:6379")

	retCode := m.Run()
	closer()

	os.Exit(retCode)

}

func TestGetInt(t *testing.T) {
	_, err := Int("123245555")
	if err == nil {
		t.Errorf("should return nil")
	}
	Set("1234565", 123)
	i, err := Int("1234565")
	if err != nil {
		t.Errorf(err.Error())
	}
	if i != 123 {
		t.Fail()
	}
	Delete("1234565")
}

func TestIncrKey(t *testing.T) {
	i, err := IncrKey("test_incr_1")
	if err != nil {
		t.Fail()
	}
	if i != 1 {
		t.Errorf("return value should equal 1")
	}
	Delete("test_incr_1")

	err = Set("test_incr_1", 10)
	if err != nil {
		t.Fail()
	}
	i, err = IncrKey("test_incr_1")
	if err != nil {
		t.Fail()
	}
	if i != 11 {
		t.Errorf("return value should equal 11")
	}
	Delete("test_incr_1")
}
