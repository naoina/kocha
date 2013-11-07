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
		c := NewErrorController(http.StatusNotFound)
		cValue := reflect.ValueOf(c)
		mValue := reflect.ValueOf(c.Get)
		controller = &cValue
		method = &mValue
		args = []reflect.Value{}
	}
	render(req, writer, controller, method, args)
}

func render(req *http.Request, writer http.ResponseWriter, controller, method *reflect.Value, args []reflect.Value) {
	request := NewRequest(req)
	response := NewResponse(writer)
	request.Body = http.MaxBytesReader(writer, request.Body, appConfig.MaxClientBodySize)
	ac := controller.Elem()
	ccValue := ac.FieldByName("Controller")
	cc := ccValue.Interface().(Controller)
	cc.Name = ac.Type().Name()
	cc.Request = request
	cc.Response = response
	cc.Params.Values = request.Form
	result := func() (result []reflect.Value) {
		defer func() {
			if err := recover(); err != nil {
				buf := make([]byte, 4096)
				runtime.Stack(buf, false)
				Log.Error("%v\n%v", err, string(buf))
				c := NewErrorController(http.StatusInternalServerError)
				c.Controller = cc
				r := c.Get()
				result = []reflect.Value{reflect.ValueOf(r)}
			}
		}()
		if err := request.ParseMultipartForm(appConfig.MaxClientBodySize); err != nil {
			panic(err)
		}
		for _, m := range appConfig.Middlewares {
			m.Before(&cc)
		}
		ccValue.Set(reflect.ValueOf(cc))
		r := method.Call(args)
		for _, m := range appConfig.Middlewares {
			m.After(&cc)
		}
		ccValue.Set(reflect.ValueOf(cc))
		return r
	}()
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
