package apiserver

import (
	"context"
	"net/http"

	"github.com/CRED-CLUB/propeller/pkg/logger"
	pushv1 "github.com/CRED-CLUB/propeller/rpc/push/v1"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc/metadata"
)

// WebSocketServerWrapper implements pushv1.PushService_ConnectServer for WebSocket
type WebSocketServerWrapper struct {
	PushServer *PushServer
	conn       *websocket.Conn
	Ctx        context.Context
	CancelFunc context.CancelFunc
}

// SetHeader ...
func (ws *WebSocketServerWrapper) SetHeader(md metadata.MD) error {
	//TODO implement me
	panic("implement me")
}

// SendHeader ...
func (ws *WebSocketServerWrapper) SendHeader(md metadata.MD) error {
	//TODO implement me
	panic("implement me")
}

// SetTrailer ...
func (ws *WebSocketServerWrapper) SetTrailer(md metadata.MD) {
	//TODO implement me
	panic("implement me")
}

// SendMsg ...
func (ws *WebSocketServerWrapper) SendMsg(m any) error {
	//TODO implement me
	panic("implement me")
}

// RecvMsg ...
func (ws *WebSocketServerWrapper) RecvMsg(m any) error {
	//TODO implement me
	panic("implement me")
}

// Send ...
func (ws *WebSocketServerWrapper) Send(resp *pushv1.ChannelResponse) error {
	return ws.conn.WriteJSON(resp)
}

// Recv ...
func (ws *WebSocketServerWrapper) Recv() (*pushv1.ChannelRequest, error) {
	var req pushv1.ChannelRequest
	err := ws.conn.ReadJSON(&req)
	if err != nil {
		ws.CancelFunc()
		return nil, err
	}
	logger.Ctx(ws.Ctx).Debugf("receive message from server: %v, %v", &req, err)
	return &req, nil
}

// Context ...
func (ws *WebSocketServerWrapper) Context() context.Context {
	return ws.Ctx
}

// WebSocketConnect handles WebSocket connections
func (ws *WebSocketServerWrapper) WebSocketConnect(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // You may want to implement proper origin checking
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Ctx(r.Context()).Errorw("Failed to upgrade WebSocket connection", "error", err)
		return
	}
	defer conn.Close()

	// add headers to metadata
	mds := metadata.New(make(map[string]string))
	for k, v := range r.Header {
		mds = metadata.Join(metadata.Pairs(k, v[0]), mds)
	}

	connCtx, connCancel := context.WithCancel(metadata.NewIncomingContext(ws.Ctx, mds))
	defer connCancel()

	connWrapper := &WebSocketServerWrapper{
		PushServer: ws.PushServer,
		conn:       conn,
		Ctx:        connCtx,
		CancelFunc: connCancel,
	}

	// Call the existing Connect method
	if err := ws.PushServer.Channel(connWrapper); err != nil {
		logger.Ctx(connCtx).Errorw("WebSocket connection error", "error", err)
	}
}
