package kocha

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
)

func handler(writer http.ResponseWriter, req *http.Request) {
	controller, method, args := dispatch(req)
	if controller == nil {
		http.NotFound(writer, req)
		return
	}
	c := controller.Elem()
	cc := c.FieldByName("Controller")
	cc.FieldByName("Name").SetString(c.Type().Name())
	cc.FieldByName("Request").Set(reflect.ValueOf(NewRequest(req)))
	cc.FieldByName("Response").Set(reflect.ValueOf(NewResponse(writer)))
	result := method.Call(args)
	result[0].Interface().(Result).Proc(writer)
}

func Run(addr string, port int) {
	if !initialized {
		log.Fatal("Uninitialized. Please call the kocha.Init() before kocha.Run()")
	}
	if addr == "" {
		addr = DefaultHttpAddr
	}
	if port == 0 {
		port = DefaultHttpPort
	}
	addr = fmt.Sprintf("%s:%d", addr, port)
	server := &http.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(handler),
	}
	fmt.Println("Listen on", addr)
	log.Fatal(server.ListenAndServe())
}
