package logging

import (
	"fmt"
	"github.com/braveghost/meteor/mode"
	"testing"
	"time"
)

func TestInitLogger(t *testing.T) {
	OpenColor()
	InitLogger(mode.ModeLocal)
	defer Sync() // flushes buffer, if any

	fmt.Println(defaultLogger)
	for {
		Debug("ddddd")
		Errorw("test err")
		time.Sleep(time.Second)
	}
}

func TestNewLogger(t *testing.T) {
	OpenColor()

	name := "xxxx"
	err := NewLogger(&Options{Path: "log", FileName: name})
	if err != nil {
		return
	}
	defer Logger("xxxx").Sync() // flushes buffer, if any

	fmt.Println(defaultLogger)
	Logger("xxxx").Debug("ddddd")
	Logger("xxxx").Errorw("test err")
}
