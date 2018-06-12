package whitelist

import (
	"os"
	"reflect"
	"testing"
)

func TestVerifyIP(t *testing.T) {
	m := map[string]bool{
		"127.0.0.1":     true,
		"127.0.0.2":     false,
		"192.168.1.1":   true,
		"192.168.1.255": true,
		"192.168.0.1":   false,
	}

	for k, v := range m {
		if VerifyIP(k) != v {
			t.Fatal(k)
		}
	}
}

func TestClearIPList(t *testing.T) {
	ClearIPList()
	m := map[string]bool{
		"127.0.0.1":     false,
		"127.0.0.2":     false,
		"192.168.1.1":   false,
		"192.168.1.255": false,
		"192.168.0.1":   false,
	}

	for k, v := range m {
		if VerifyIP(k) != v {
			t.Fatal(k)
		}
	}
}

func TestRegisterIP(t *testing.T) {
	RegisterIP("159.56.25.14")

	if !VerifyIP("159.56.25.14") {
		t.Fail()
	}
}

func TestRemoveIP(t *testing.T) {
	RegisterIP("159.56.25.14")

	if !VerifyIP("159.56.25.14") {
		t.Fail()
	}

	RemoveIP("159.56.25.14")
	if VerifyIP("159.56.25.14") {
		t.Fail()
	}
}

func TestIPList(t *testing.T) {
	ClearIPList()
	m := []string{"124.4.59.24", "58.57.1.*"}

	for _, ip := range m {
		RegisterIP(ip)
	}

	if !reflect.DeepEqual(m, IPList()) {
		t.Fail()
	}
}

func TestMain(m *testing.M) {
	Setup([]string{"127.0.0.1", "192.168.1.*"})

	retCode := m.Run()
	os.Exit(retCode)

}
