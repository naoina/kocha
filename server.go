package kocha

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"runtime"

	"github.com/naoina/miyabi"
)

func handler(writer http.ResponseWriter, req *http.Request) {
	controller, method, args := appConfig.router.dispatch(req)
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
	request := newRequest(req)
	response := newResponse(writer)
	var (
		cc     *Controller
		result []reflect.Value
	)
	defer func() {
		defer func() {
			if err := recover(); err != nil {
				logStackAndError(err)
				response.StatusCode = http.StatusInternalServerError
				http.Error(response, http.StatusText(response.StatusCode), response.StatusCode)
			}
		}()
		if err := recover(); err != nil {
			logStackAndError(err)
			c := NewErrorController(http.StatusInternalServerError)
			if cc == nil {
				cc = &Controller{}
				cc.Request = request
				cc.Response = response
			}
			c.Controller = cc
			r := c.Get()
			result = []reflect.Value{reflect.ValueOf(r)}
		}
		for _, m := range appConfig.Middlewares {
			m.After(cc)
		}
		response.WriteHeader(response.StatusCode)
		result[0].Interface().(Result).Proc(response)
	}()
	request.Body = http.MaxBytesReader(writer, request.Body, appConfig.MaxClientBodySize)
	ac := controller.Elem()
	ccValue := ac.FieldByName("Controller")
	switch c := ccValue.Interface().(type) {
	case Controller:
		cc = &c
	case *Controller:
		cc = &Controller{}
		ccValue.Set(reflect.ValueOf(cc))
		ccValue = ccValue.Elem()
	default:
		panic(fmt.Errorf("BUG: Controller field must be struct of %T or that pointer, but %T", cc, c))
	}
	if err := request.ParseMultipartForm(appConfig.MaxClientBodySize); err != nil && err != http.ErrNotMultipart {
		panic(err)
	}
	cc.Name = ac.Type().Name()
	cc.Layout = appConfig.DefaultLayout
	cc.Context = Context{}
	cc.Request = request
	cc.Response = response
	cc.Params = newParams(cc, request.Form, "")
	for _, m := range appConfig.Middlewares {
		m.Before(cc)
	}
	ccValue.Set(reflect.ValueOf(*cc))
	result = method.Call(args)
}

// Run starts Kocha app.
//
// addr is string of address for bind.
func Run(addr string) {
	if !initialized {
		log.Fatalln("Uninitialized. Please call the kocha.Init() before kocha.Run()")
	}
	if addr == "" {
		addr = DefaultHttpAddr
	}
	pid := os.Getpid()
	miyabi.ServerState = func(state miyabi.State) {
		switch state {
		case miyabi.StateStart:
			fmt.Printf("Listening on %s\n", addr)
			fmt.Printf("Server PID: %d\n", pid)
		case miyabi.StateRestart:
			Log.Warn("graceful restarted")
		case miyabi.StateShutdown:
			Log.Warn("graceful shutdown")
		}
	}
	server := &miyabi.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(handler),
	}
	server.ListenAndServe()
}

func logStackAndError(err interface{}) {
	buf := make([]byte, 4096)
	runtime.Stack(buf, false)
	Log.Error("%v\n%v", err, string(buf))
}
