package kocha

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"syscall"
)

const fdKey = "KOCHA_FD"

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
	request := NewRequest(req)
	response := NewResponse(writer)
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
	cc.Name = ac.Type().Name()
	cc.Layout = appConfig.DefaultLayout
	cc.Context = Context{}
	cc.Request = request
	cc.Response = response
	cc.Params = newParams(cc, request.Form, "")
	if err := request.ParseMultipartForm(appConfig.MaxClientBodySize); err != nil && err != http.ErrNotMultipart {
		panic(err)
	}
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
	l, reloaded := serverListener(addr)
	listener := &waitableListener{
		Listener: l,
		wg:       &sync.WaitGroup{},
	}
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP)
	go func() {
		switch <-c {
		case syscall.SIGHUP:
			pid := gracefulRestart(listener)
			Log.Warn("graceful restarted. new pid: %d", pid)
			if err := listener.Close(); err != nil {
				panic(err)
			}
		}
	}()
	server := &http.Server{
		Handler: http.HandlerFunc(handler),
	}
	if !reloaded {
		fmt.Printf("Listen on %s, pid %d\n", addr, os.Getpid())
	}
	server.Serve(listener)
	listener.wg.Wait()
}

func gracefulRestart(listener *waitableListener) (pid int) {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fdValue := reflect.Indirect(reflect.Indirect(reflect.ValueOf(listener.Listener)).FieldByName("fd"))
	sysfd := uintptr(fdValue.FieldByName("sysfd").Int())
	proc, err := os.StartProcess(os.Args[0], os.Args, &os.ProcAttr{
		Dir:   pwd,
		Env:   append(os.Environ(), fmt.Sprintf("%s=%d", fdKey, sysfd)),
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr, os.NewFile(sysfd, "sysfile")},
	})
	if err != nil {
		panic(err)
	}
	return proc.Pid
}

func serverListener(addr string) (listener net.Listener, reloaded bool) {
	if fdStr := os.Getenv(fdKey); fdStr != "" {
		fd, err := strconv.Atoi(fdStr)
		if err != nil {
			panic(err)
		}
		file := os.NewFile(uintptr(fd), "listen socket")
		l, err := net.FileListener(file)
		if err != nil {
			panic(err)
		}
		return l, true
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		panic(err)
	}
	l, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		panic(err)
	}
	return l, false
}

type waitableListener struct {
	net.Listener
	wg *sync.WaitGroup
}

func (l *waitableListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return conn, err
	}
	l.wg.Add(1)
	return &waitableConn{Conn: conn, wg: l.wg}, nil
}

type waitableConn struct {
	net.Conn
	wg *sync.WaitGroup
}

func (c *waitableConn) Close() error {
	if err := c.Conn.Close(); err != nil {
		return err
	}
	c.wg.Done()
	return nil
}

func logStackAndError(err interface{}) {
	buf := make([]byte, 4096)
	runtime.Stack(buf, false)
	Log.Error("%v\n%v", err, string(buf))
}
