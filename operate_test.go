package joker

import (
	"context"
	"fmt"
	"testing"
)

func TestDebugwc(t *testing.T) {
	ct := context.Background()
	ct = context.WithValue(ct, requestIdKey, 123)
	fmt.Println(GetRequestId(ct))

	Debugwc("test", ct)

	ct1 := context.Background()
	ct = context.WithValue(ct1, "xxx", 123)
	fmt.Println(GetRequestId(ct))
	Debugwc("test", ct)

}
