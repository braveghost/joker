package joker

import (
	"fmt"
	"github.com/braveghost/meteor/mode"
	"testing"
)

func TestInitLogger(t *testing.T) {

	InitLogger(mode.ModeLocal)
	defer Sync() // flushes buffer, if any

	fmt.Println(defaultLogger)
	Debug("ddddd")
	Errorw("test err")
}


func TestNewLogger(t *testing.T) {

	name := ""
	err := NewLogger(&LoggingConf{Path:"log", Name:name})
	if err != nil {
		return
	}
	defer GetLogger(name).Sync() // flushes buffer, if any

	fmt.Println(defaultLogger)
	GetLogger(name).Debug("ddddd")
	GetLogger(name).Errorw("test err")
}
