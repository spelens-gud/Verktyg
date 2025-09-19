package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	agent "github.com/SkyAPM/go2sky/reporter/grpc/language-agent"
	"github.com/gin-gonic/gin/binding"
	"google.golang.org/grpc"

	"git.bestfulfill.tech/devops/go-core/kits/klog/logger"
)

var (
	addr       = flag.String("addr", "0.0.0.0:20202", "listen addr")
	reportAddr = flag.String("report", "localhost:11800", "report grpc addr")
)

func main() {
	flag.Parse()
	if len(*reportAddr) == 0 {
		panic("invalid report rpc addr")
	}

	var (
		err error
		l   net.Listener

		listenAddr = *addr
	)

	if strings.HasSuffix(listenAddr, ".sock") {
		_ = os.Remove(listenAddr)
		l, err = net.Listen("unix", listenAddr)
	} else {
		l, err = net.Listen("tcp", listenAddr)
	}

	if err != nil {
		panic(err)
	}
	defer l.Close()

	buf := make(chan *agent.SegmentObject, 2<<10)
	ctx, cf := context.WithCancel(context.Background())
	go handle(ctx, buf)
	go func() {
		if err := (&http.Server{
			ConnState: func(conn net.Conn, state http.ConnState) {
				logger.FromContext(ctx).Infof("connection state change [ %s ] state: %s", conn.RemoteAddr().String(), state.String())
			},
			Handler:      receive(buf),
			ReadTimeout:  time.Second * 10,
			WriteTimeout: time.Second * 10,
			IdleTimeout:  time.Second * 120,
			ErrorLog:     logger.StandardLogger(),
			BaseContext: func(listener net.Listener) context.Context {
				return ctx
			},
		}).Serve(l); err != nil {
			logger.FromContext(ctx).Errorf("server run error: %v", err)
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	logger.FromBackground().Infof("agent start")
	<-c
	cf()
	time.Sleep(time.Second)
}

func handle(ctx context.Context, handleChan chan *agent.SegmentObject) {
	conn, err := grpc.Dial(*reportAddr, grpc.WithInsecure())
	if err != nil {
		panic(fmt.Errorf("dial grpc connect error: %v", err))
	}
	defer conn.Close()
	client := agent.NewTraceSegmentReportServiceClient(conn)
	logger.FromBackground().Infof("agent stream start")

StreamLoop:
	for {
		stream, err := client.Collect(ctx)
		if err != nil {
			logger.FromBackground().Errorf("collect stream error: %v", err)
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
			continue StreamLoop
		}

		for {
			select {
			case <-ctx.Done():
				_ = stream.CloseSend()
				return
			case obj, ok := <-handleChan:
				if !ok {
					_ = stream.CloseSend()
					return
				}
				if err := stream.Send(obj); err != nil {
					_ = stream.CloseSend()
					logger.FromBackground().Errorf("send stream buf error: %v", err)
					continue StreamLoop
				}
			}
		}
	}
}

func receive(buf chan *agent.SegmentObject) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		lg := logger.FromContext(ctx)

		defer func() {
			if e := recover(); e != nil {
				lg.WithTag("SERVER_PANIC").Errorf("server recover from panic: %+v", e)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
		}()

		var tmp agent.SegmentObject
		err := binding.Default(request.Method, request.Header.Get("Content-Type")).Bind(request, &tmp)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			lg.Errorf("parse request error: %v", err)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}
		if len(tmp.Spans) == 0 || len(tmp.TraceId) == 0 {
			lg.Errorf("invalid request %v", tmp)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		select {
		case <-ctx.Done():
			lg.Errorf("context closed: %v", ctx.Err())
			writer.WriteHeader(http.StatusInternalServerError)
			return
		case buf <- &tmp:
			lg.Infof("receive traceID [ %s ] segment with %d span success", tmp.TraceId, len(tmp.Spans))
			writer.WriteHeader(http.StatusOK)
		default:
			lg.Warnf("buff overflow,segment drop: %v", tmp)
			writer.WriteHeader(http.StatusTooManyRequests)
		}
	}
}
