package joker

import (
	"context"
	"fmt"
	"testing"
)

func TestGetRequestId(t *testing.T) {
	ct := context.Background()
	ct = context.WithValue(ct, traceIdKey, 123)
	fmt.Println(GetTraceId(ct))

	ct1 := context.Background()
	ct = context.WithValue(ct1, "xxx", 123)
	fmt.Println(GetTraceId(ct))

}
