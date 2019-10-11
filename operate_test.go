package joker

import (
	"fmt"
	"github.com/braveghost/meteor/mode"
	"testing"
)

func TestDebugwc(t *testing.T) {
	//ct := context.Background()
	//ct = context.WithValue(ct, requestIdKey, 123)
	//fmt.Println(GetRequestId(ct))
	//
	//Debugwc("test", ct)
	//
	//ct1 := context.Background()
	//ct = context.WithValue(ct1, "xxx", 123)
	//fmt.Println(GetRequestId(ct))

	GetLogger("tt", mode.ModePro)
	defer Sync() // flushes buffer, if any

	fmt.Println(defaultLogger)
	Debug("ddddd")
	Errorw("test err")
}
