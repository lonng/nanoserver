package room

import (
	"math/rand"
	"sync"
	"time"

	"github.com/lonng/nanoserver/db"
)

const (
	roomNoLen = 6
)

type Number string
type numberManager struct {
	lock sync.Mutex
}

var rn *numberManager
var numbers = [...]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}

func init() {
	rn = &numberManager{}

	rand.Seed(time.Now().Unix())
}

func (rn *numberManager) next() Number {
	no := make([]byte, roomNoLen)
	rn.lock.Lock()
	defer rn.lock.Unlock()

	for {
		for i := 0; i < roomNoLen; i++ {
			no[i] = numbers[rand.Intn(10)]
		}
		temp := Number(no)
		dn := string(no)
		if !db.DeskNumberExists(dn) {
			return temp
		}

	}
}

func (rn *numberManager) remove(no Number) {
	//rn.noPool.Remove(string(no))
}

func Next() Number {
	return rn.next()
}

func (n Number) String() string {
	return string(n)
}
