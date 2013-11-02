package kocha

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime"
)

func handler(writer http.ResponseWriter, req *http.Request) {
	controller, method, args := dispatch(req)
	if controller == nil {
		http.NotFound(writer, req)
		return
	}
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			runtime.Stack(buf, false)
			log.Print(err)
			log.Print(string(buf))
			http.Error(writer, "500 Internal Server Error", 500)
		}
	}()
	request := NewRequest(req)
	response := NewResponse(writer)
	request.Body = http.MaxBytesReader(writer, request.Body, maxClientBodySize)
	if err := request.ParseMultipartForm(maxClientBodySize); err != nil {
		panic(err)
	}
	c := controller.Elem()
	cc := c.FieldByName("Controller")
	cc.FieldByName("Name").SetString(c.Type().Name())
	cc.FieldByName("Request").Set(reflect.ValueOf(request))
	cc.FieldByName("Response").Set(reflect.ValueOf(response))
	cc.FieldByName("Params").FieldByName("Values").Set(reflect.ValueOf(request.Form))
	result := method.Call(args)
	response.Header().Set("Content-Type", response.ContentType+"; charset=utf-8")
	response.WriteHeader(response.StatusCode)
	result[0].Interface().(Result).Proc(response)
}

func Run(addr string, port int) {
	if !initialized {
		log.Fatalln("Uninitialized. Please call the kocha.Init() before kocha.Run()")
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
