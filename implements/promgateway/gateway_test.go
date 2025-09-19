package promgateway

import (
	"fmt"
	"io"
	"net/http"
	"testing"
)

func TestGateway(t *testing.T) {
	go func() {
		_ = http.ListenAndServe(":80", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			fmt.Printf("%v\n", request.Header)
			b, _ := io.ReadAll(request.Body)
			fmt.Printf("%s\n", b)
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte("OK"))

		}))
	}()

	gateway := GatewayConfig{
		Job:             "test",
		GatewayUrl:      "http://localhost",
		IntervalSeconds: 0,
		EnableLog:       true,
	}.NewTransport()
	defer gateway.Stop()
	gateway.StartDaemon()
}
