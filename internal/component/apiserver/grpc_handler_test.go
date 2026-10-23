package apiserver

import (
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/CRED-CLUB/propeller/internal/config"
	"github.com/CRED-CLUB/propeller/pkg/logger"
	pushv1 "github.com/CRED-CLUB/propeller/rpc/push/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func init() {
	_, _ = logger.NewLogger("dev", nil, nil)
}

type mockChannelServer struct {
	ctx        context.Context
	recvQueue  chan *pushv1.ChannelRequest
	recvErr    error
	recvNotify chan struct{}
	mu         sync.Mutex
	sent       []*pushv1.ChannelResponse
}

func newMockChannelServer(ctx context.Context) *mockChannelServer {
	return &mockChannelServer{
		ctx:       ctx,
		recvQueue: make(chan *pushv1.ChannelRequest, 10),
	}
}

func (m *mockChannelServer) Context() context.Context {
	return m.ctx
}

func (m *mockChannelServer) Recv() (*pushv1.ChannelRequest, error) {
	if m.recvErr != nil {
		return nil, m.recvErr
	}
	select {
	case req, ok := <-m.recvQueue:
		if !ok {
			return nil, io.EOF
		}
		if m.recvNotify != nil {
			select {
			case m.recvNotify <- struct{}{}:
			default:
			}
		}
		return req, nil
	case <-m.ctx.Done():
		return nil, m.ctx.Err()
	}
}

func (m *mockChannelServer) Send(resp *pushv1.ChannelResponse) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sent = append(m.sent, resp)
	return nil
}

func (m *mockChannelServer) SetHeader(metadata.MD) error { return nil }
func (m *mockChannelServer) SendHeader(metadata.MD) error { return nil }
func (m *mockChannelServer) SetTrailer(metadata.MD) {}
func (m *mockChannelServer) SendMsg(any) error { return nil }
func (m *mockChannelServer) RecvMsg(any) error { return nil }

func TestValidateAndGetClientAndDeviceDetails(t *testing.T) {
	cfg := config.Config{
		ClientHeader:           "x-client-id",
		DeviceHeader:           "x-device-id",
		EnableDeviceSupport:    true,
		DeviceAttributeHeaders: []string{"x-os-version"},
	}
	server := NewPushServer(nil, cfg)

	t.Run("valid metadata with client and device", func(t *testing.T) {
		md := metadata.New(map[string]string{
			"x-client-id":  "client-123",
			"x-device-id":  "device-abc",
			"x-os-version": "14.2",
		})
		ctx := metadata.NewIncomingContext(context.Background(), md)

		clientID, device, err := server.validateAndGetClientAndDeviceDetails(ctx)
		require.NoError(t, err)
		assert.Equal(t, "client-123", clientID)
		require.NotNil(t, device)
		assert.Equal(t, "device-abc", device.ID)
		assert.Equal(t, "14.2", device.Attributes["x-os-version"])
	})

	t.Run("missing metadata", func(t *testing.T) {
		ctx := context.Background()
		_, _, err := server.validateAndGetClientAndDeviceDetails(ctx)
		require.Error(t, err)
	})

	t.Run("missing client header", func(t *testing.T) {
		md := metadata.New(map[string]string{
			"x-device-id": "device-abc",
		})
		ctx := metadata.NewIncomingContext(context.Background(), md)
		_, _, err := server.validateAndGetClientAndDeviceDetails(ctx)
		require.Error(t, err)
	})

	t.Run("missing device header when device support enabled", func(t *testing.T) {
		md := metadata.New(map[string]string{
			"x-client-id": "client-123",
		})
		ctx := metadata.NewIncomingContext(context.Background(), md)
		_, _, err := server.validateAndGetClientAndDeviceDetails(ctx)
		require.Error(t, err)
	})

	t.Run("device support disabled", func(t *testing.T) {
		cfgNoDevice := config.Config{
			ClientHeader:        "x-client-id",
			EnableDeviceSupport: false,
		}
		s := NewPushServer(nil, cfgNoDevice)
		md := metadata.New(map[string]string{
			"x-client-id": "client-123",
		})
		ctx := metadata.NewIncomingContext(context.Background(), md)
		clientID, device, err := s.validateAndGetClientAndDeviceDetails(ctx)
		require.NoError(t, err)
		assert.Equal(t, "client-123", clientID)
		require.NotNil(t, device)
		assert.Empty(t, device.ID)
	})
}

func TestHandleReceivedPayload_NilSafety(t *testing.T) {
	server := NewPushServer(nil, config.Config{})
	mockStream := newMockChannelServer(context.Background())

	assert.NotPanics(t, func() {
		server.HandleReceivedPayload(context.Background(), mockStream, nil, nil)
	})

	assert.NotPanics(t, func() {
		server.HandleReceivedPayload(context.Background(), mockStream, &pushv1.ChannelRequest{}, nil)
	})
}

func TestReceiveLoop_EOFExitAndCloseChannel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	mockStream := newMockChannelServer(ctx)
	close(mockStream.recvQueue) // triggers io.EOF on Recv()

	rc := make(chan *pushv1.ChannelRequest)
	done := make(chan struct{})

	go func() {
		receiveLoop(ctx, rc, mockStream)
		close(done)
	}()

	select {
	case <-done:
		// receiveLoop exited cleanly
	case <-time.After(500 * time.Millisecond):
		t.Fatal("receiveLoop did not terminate on EOF")
	}

	// Verify rc channel is closed
	_, ok := <-rc
	assert.False(t, ok, "expected rc channel to be closed by receiveLoop")
}

func TestReceiveLoop_ContextCancellationWhileBlocked(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockStream := newMockChannelServer(context.Background())
	mockStream.recvNotify = make(chan struct{}, 1)

	rc := make(chan *pushv1.ChannelRequest) // unbuffered channel, no reader
	done := make(chan struct{})

	mockStream.recvQueue <- &pushv1.ChannelRequest{}

	go func() {
		receiveLoop(ctx, rc, mockStream)
		close(done)
	}()

	// Wait until Recv() has dispatched the item and receiveLoop is blocked on `rc <- req`
	select {
	case <-mockStream.recvNotify:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for Recv() dispatch")
	}

	// Cancel context while receiveLoop is blocked on unbuffered send
	cancel()

	select {
	case <-done:
		// Passed: cleanly exited on cancellation without blocking on rc
	case <-time.After(500 * time.Millisecond):
		t.Fatal("receiveLoop remained blocked after context cancellation")
	}

	// Verify rc is closed by defer
	_, ok := <-rc
	assert.False(t, ok, "expected rc channel to be closed")
}
