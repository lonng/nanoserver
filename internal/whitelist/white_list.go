package whitelist

import (
	"regexp"
	"sync"
)

var (
	lock sync.RWMutex
	ips  = map[string]*regexp.Regexp{}
)

func Setup(list []string) error {
	lock.Lock()
	defer lock.Unlock()

	for _, ip := range list {
		re, err := regexp.Compile(ip)
		if err != nil {
			return err
		}
		ips[ip] = re
	}

	return nil
}

//VerifyIP check the ip is a legal ip or not
func VerifyIP(ip string) bool {
	lock.RLock()
	defer lock.RUnlock()

	for _, r := range ips {
		if r.MatchString(ip) {
			return true
		}
	}
	return false
}

func RegisterIP(ip string) error {
	lock.Lock()
	defer lock.Unlock()

	_, ok := ips[ip]
	if ok {
		return nil
	}

	re, err := regexp.Compile(ip)
	if err != nil {
		return err
	}
	ips[ip] = re
	return nil
}

func RemoveIP(ip string) {
	lock.Lock()
	defer lock.Unlock()

	delete(ips, ip)
}

func IPList() []string {
	lock.RLock()
	defer lock.RUnlock()

	list := []string{}
	for ip := range ips {
		list = append(list, ip)
	}

	return list
}

func ClearIPList() {
	lock.Lock()
	defer lock.Unlock()

	ips = map[string]*regexp.Regexp{}
}
