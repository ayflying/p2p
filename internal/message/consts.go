package message

import (
	v1 "github.com/ayflying/p2p/api/p2p/v1"
	"github.com/ayflying/p2p/internal/message/http"
	"github.com/ayflying/p2p/internal/message/proxy"
	"github.com/ayflying/p2p/internal/message/proxyAck"
)

type P2PMessage interface {
	Message(msg *v1.Message) (err error)
}

var Run = map[string]P2PMessage{
	"proxy":     proxy.New,
	"proxy_ack": proxyAck.New,
	"http":      http.New,
}
