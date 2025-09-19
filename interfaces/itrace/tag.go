package itrace

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/opentracing/opentracing-go/ext"
)

const (
	TagServiceCode = "service.code"
	TagRpcMethod   = "rpc.method"
	TagSpanType    = "span.type"
	TagComponentID = "component.id"

	TagMessageBus           = "message_bus"
	TagMessageBusType       = TagMessageBus + ".type"
	TagMessageBusClient     = TagMessageBus + ".client"
	TagMessageBusClientType = TagMessageBus + ".client.type"
	TagMessageBusAddress    = TagMessageBus + ".address"
	TagMessagePartition     = TagMessageBus + ".partition"
	TagMessageOffset        = TagMessageBus + ".offset"

	TagDBStatementType = "db.statement.type"

	SpanTypeExit  = "exit"
	SpanTypeEntry = "entry"
	SpanTypeLocal = "local"
)

// nolint
var unknownErr = errors.New("unknown error")

func setHttpTag(span Span, req *http.Request) {
	span.Tag(string(ext.HTTPMethod), req.Method)
	span.Tag(string(ext.HTTPUrl), req.URL.Path)
	span.SetPeer(req.Host)
}

func SetHttpClientTag(span Span, req *http.Request) {
	span.SetComponent(ComponentGoHttpClient)
	span.Tag(string(ext.SpanKind), string(ext.SpanKindRPCClientEnum))
	setHttpTag(span, req)
}

func SetHttpServerTag(span Span, req *http.Request) {
	span.SetComponent(ComponentGoHttpServer)
	span.Tag(string(ext.SpanKind), string(ext.SpanKindRPCServerEnum))
	setHttpTag(span, req)
}

func SetPeerIPTag(span Span, ip net.IP) {
	switch {
	case ip.To4() != nil:
		span.Tag(string(ext.PeerHostIPv4), ip.String())
	case ip.To16() != nil:
		span.Tag(string(ext.PeerHostIPv6), ip.String())
	}
}

func SetNetPeerTag(span Span, addr net.Addr) {
	switch addr := addr.(type) {
	case *net.TCPAddr:
		SetPeerIPTag(span, addr.IP)
		span.Tag(string(ext.PeerPort), strconv.Itoa(addr.Port))
	case *net.IPAddr:
		SetPeerIPTag(span, addr.IP)
	case *net.UDPAddr:
		SetPeerIPTag(span, addr.IP)
	}
}

func SetHttpStatusTag(span Span, status int, err error) {
	span.Tag(string(ext.HTTPStatusCode), strconv.Itoa(status))
	switch {
	case status >= 400:
		statusText := http.StatusText(status)
		if len(statusText) == 0 {
			statusText = "Unknown Http Status"
		}
		if err == nil {
			err = unknownErr
		}
		span.Error(fmt.Errorf("%s [ %d ]: %s", statusText, status, err.Error()), nil)
	case status < 0:
		if err == nil {
			err = unknownErr
		}
		span.Error(fmt.Errorf("Request Error: "+err.Error()), nil)
	}
}

func SetServiceCodeTag(span Span, code int) {
	span.Tag(TagServiceCode, strconv.Itoa(code))
}

func SetRpcClientTag(span Span) {
	span.SetComponent(ComponentRpc)
	span.Tag(string(ext.SpanKind), string(ext.SpanKindRPCClientEnum))
}

func SetRpcServerTag(span Span) {
	span.SetComponent(ComponentRpc)
	span.Tag(string(ext.SpanKind), string(ext.SpanKindRPCServerEnum))
}

func SetRpcServiceTag(span Span, serviceName string) {
	span.Tag(string(ext.PeerService), serviceName)
}

func SetRpcAddressTag(span Span, address string) {
	span.Tag(string(ext.PeerAddress), address)
}

func SetRpcMethodTag(span Span, method string) {
	span.Tag(TagRpcMethod, method)
}

func SetMqMessageTag(sp Span, partition, offset int) {
	sp.Tag(TagMessagePartition, strconv.Itoa(partition))
	sp.Tag(TagMessageOffset, strconv.Itoa(offset))
}

func SetMqTag(sp Span, address, topic, mqType, client, clientType string, isProduce bool) {
	sp.Tag(TagMessageBusType, mqType)
	sp.Tag(TagMessageBusClient, client)
	sp.Tag(TagMessageBusClientType, clientType)
	sp.Tag(TagMessageBusAddress, address)
	sp.Tag(string(ext.MessageBusDestination), topic)
	sp.SetLayer(SpanLayerMQ)
	if isProduce {
		sp.Tag(string(ext.SpanKind), string(ext.SpanKindProducerEnum))
		sp.SetComponent(ComponentKafkaProducer)
	} else {
		sp.Tag(string(ext.SpanKind), string(ext.SpanKindConsumerEnum))
		sp.SetComponent(ComponentKafkaConsumer)
	}
}
