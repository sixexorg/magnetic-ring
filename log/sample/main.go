package main

import (
	"github.com/sixexorg/magnetic-ring/log"
)

type Tem struct {
	FieldA string
	FieldB int64
}

func ab()  {
	//log.Lvl(log.LvlError)
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlInfo,log.Must.FileHandler("errors.json", log.JsonFormat())))
	log.Info("okgoogog","a","bbbs")
	log.Error("okgoogog","a","bbbs")
	log.Debug("okgoogog","a","bbbs")
	log.Warn("ookkksss","a","warnwarn")
}

func main() {



	ab()
	//fmt.Printf("a=%d\n",8)
	//srvlog := log.New("module", "app/server")
	//
	//
	//// flexible configuration
	//srvlog.SetHandler(log.MultiHandler(
	//	log.StderrHandler,
	//	log.LvlFilterHandler(
	//		log.LvlError,
	//		log.Must.FileHandler("errors.json", log.JsonFormat()))))
	//
	//srvlog.Info("where msg from","data mark","data")
	//
	//t := new(Tem)
	//t.FieldA="string field"
	//t.FieldB=123456
	//
	//srvlog.Error("what msg do for",t,t)

}
