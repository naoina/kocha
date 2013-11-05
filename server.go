package kocha

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime"
)

func handler(writer http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			runtime.Stack(buf, false)
			Log.Error("%v\n%v", err, string(buf))
			controller, method, args := errorDispatch(http.StatusInternalServerError)
			render(req, writer, controller, method, args)
		}
	}()
	controller, method, args := dispatch(req)
	if controller == nil {
		controller, method, args = errorDispatch(http.StatusNotFound)
	}
	render(req, writer, controller, method, args)
}

func render(req *http.Request, writer http.ResponseWriter, controller, method *reflect.Value, args []reflect.Value) {
	request := NewRequest(req)
	response := NewResponse(writer)
	request.Body = http.MaxBytesReader(writer, request.Body, maxClientBodySize)
	if err := request.ParseMultipartForm(maxClientBodySize); err != nil {
		panic(err)
	}
	ac := controller.Elem()
	ccValue := ac.FieldByName("Controller")
	cc := ccValue.Interface().(Controller)
	cc.Name = ac.Type().Name()
	cc.Request = request
	cc.Response = response
	cc.Params.Values = request.Form
	ccValue.Set(reflect.ValueOf(cc))
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
