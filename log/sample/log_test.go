package main

import (
	"fmt"
	"github.com/sixexorg/magnetic-ring/log"
	"testing"
)

func TestLog(t *testing.T)  {
	//log.Root().SetHandler())
	log.Root().SetHandler(log.MultiHandler(log.LvlFilterHandler(log.LvlInfo,log.StdoutHandler),log.LvlFilterHandler(log.LvlInfo,log.Must.FileHandler("magnetic.log", log.TerminalFormat(true)))))



	log.Info(fmt.Sprintf("a=%d",8888))
	log.Error("okgoogog","a","bbbs")
	log.Debug("okgoogog","a","bbbs")
	log.Warn("ookkksss","a","warnwarn")
}