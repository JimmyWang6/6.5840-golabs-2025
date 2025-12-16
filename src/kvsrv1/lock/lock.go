package lock

import (
	"log"

	"6.5840/kvsrv1/rpc"
	"6.5840/kvtest1"
)

type Lock struct {
	// IKVClerk is a go interface for k/v clerks: the interface hides
	// the specific Clerk type of ck but promises that ck supports
	// Put and Get.  The tester passes the clerk in when calling
	// MakeLock().
	ck      kvtest.IKVClerk
	version rpc.Tversion
	l       string
	// You may add code here
}

// The tester calls MakeLock() and passes in a k/v clerk; your code can
// perform a Put or Get by calling lk.ck.Put() or lk.ck.Get().
//
// Use l as the key to store the "lock state" (you would have to decide
// precisely what the lock state is).
func MakeLock(ck kvtest.IKVClerk, l string) *Lock {
	lk := &Lock{ck: ck, l: l, version: 0}
	// You may add code here
	return lk
}

func (lk *Lock) Acquire() {
	for {
		res, version, err := lk.ck.Get(lk.l)
		if err == rpc.OK {
			if res == "acquire" {
				// other people hold the lock
				continue
			} else {
				newVersion := version
				e := lk.ck.Put(lk.l, "acquire", newVersion)
				if e == rpc.OK {
					lk.version = newVersion
					break
				}
			}
		} else {
			// first time acquire the lock
			e := lk.ck.Put(lk.l, "acquire", 0)
			if e == rpc.OK {
				break
			}
		}
	}
	// Your code here
}

func (lk *Lock) Release() {
	err := lk.ck.Put(lk.l, "", lk.version+1)
	log.Printf("release lock err: %v", err)
}
