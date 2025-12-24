package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	// 命令行参数，用于指定A和C服务的WebSocket URL
	aUrl = flag.String("a_url", "ws://localhost:8081/v3/ws.MessageWebSocket/:stream_key", "A service WebSocket URL (e.g., ws://localhost:8081/ws)")
	cUrl = flag.String("c_url", "ws://127.0.0.1:9222/devtools/browser/3747f5a4-1d55-479b-a54f-42023151d19f", "C service WebSocket URL (e.g., ws://localhost:8082/ws)")
)

// ServiceBridge represents the B service that bridges A and C
type ServiceBridge struct {
	aConn     *websocket.Conn
	cConn     *websocket.Conn
	connMutex sync.Mutex
	stopChan  chan struct{} // 用于通知所有协程停止
}

// NewServiceBridge creates a new instance of the bridge
func NewServiceBridge() *ServiceBridge {
	return &ServiceBridge{
		stopChan: make(chan struct{}),
	}
}

// connectToService connects to a given WebSocket URL with retries
func (sb *ServiceBridge) connectToService(url string, timeout time.Duration) (*websocket.Conn, error) {
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = timeout

	// 重试连接
	for {
		select {
		case <-sb.stopChan:
			return nil, context.Canceled
		default:
		}

		log.Printf("Attempting to connect to service at %s...", url)
		conn, _, err := dialer.Dial(url, nil)
		if err != nil {
			log.Printf("Failed to connect to service %s: %v. Retrying in 5 seconds...", url, err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Printf("Successfully connected to service at %s.", url)
		return conn, nil
	}
}

// run starts the bridging process
func (sb *ServiceBridge) run(ctx context.Context) error {
	// 并发地连接到A和C
	var wg sync.WaitGroup
	var aConn, cConn *websocket.Conn
	var aErr, cErr error

	wg.Add(2)

	// Connect to A
	go func() {
		defer wg.Done()
		aConn, aErr = sb.connectToService(*aUrl, 10*time.Second)
		if aErr != nil {
			log.Printf("Failed to connect to A permanently: %v", aErr)
			close(sb.stopChan) // 任一连接失败都应触发停止
		}
	}()

	// Connect to C
	go func() {
		defer wg.Done()
		cConn, cErr = sb.connectToService(*cUrl, 10*time.Second)
		if cErr != nil {
			log.Printf("Failed to connect to C permanently: %v", cErr)
			close(sb.stopChan) // 任一连接失败都应触发停止
		}
	}()

	wg.Wait()

	// 检查是否有连接失败
	if aErr != nil || cErr != nil {
		return fmt.Errorf("failed to connect to one or both services: A: %v, C: %v", aErr, cErr)
	}

	// 存储连接
	sb.connMutex.Lock()
	sb.aConn = aConn
	sb.cConn = cConn
	sb.connMutex.Unlock()

	// 启动双向消息转发
	go sb.forwardMessages(aConn, cConn, "A->C")
	go sb.forwardMessages(cConn, aConn, "C->A")

	// 等待上下文取消信号或停止信号
	select {
	case <-ctx.Done():
		log.Println("Context cancelled, stopping bridge...")
	case <-sb.stopChan: // 如果连接失败也会触发
		log.Println("Stop signal received, stopping bridge...")
	}

	// 清理资源
	sb.closeConnections()
	return nil
}

// forwardMessages reads messages from src and writes them to dst
func (sb *ServiceBridge) forwardMessages(src, dst *websocket.Conn, direction string) {
	defer func() {
		// 如果转发协程出错退出，通知其他协程停止
		select {
		case sb.stopChan <- struct{}{}:
		default:
		}
	}()

	for {
		select {
		case <-sb.stopChan:
			return
		default:
		}

		mt, message, err := src.ReadMessage()
		if err != nil {
			log.Printf("Error reading from %s (%s): %v", sb.getConnName(src), direction, err)
			return
		}

		err = dst.WriteMessage(mt, message)
		if err != nil {
			log.Printf("Error writing to %s (%s): %v", sb.getConnName(dst), direction, err)
			return
		}
		log.Printf("[%s] Forwarded message: %s", direction, string(message))
	}
}

// getConnName returns a simple name for logging purposes
func (sb *ServiceBridge) getConnName(conn *websocket.Conn) string {
	if conn == sb.aConn {
		return "A"
	} else if conn == sb.cConn {
		return "C"
	}
	return "Unknown"
}

// closeConnections safely closes both WebSocket connections
func (sb *ServiceBridge) closeConnections() {
	sb.connMutex.Lock()
	defer sb.connMutex.Unlock()

	if sb.aConn != nil {
		if err := sb.aConn.Close(); err != nil {
			log.Printf("Error closing connection to A: %v", err)
		}
	}
	if sb.cConn != nil {
		if err := sb.cConn.Close(); err != nil {
			log.Printf("Error closing connection to C: %v", err)
		}
	}
}

func main() {
	flag.Parse()

	bridge := NewServiceBridge()

	// 创建一个可取消的上下文，用于优雅地停止整个桥接过程
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Printf("B service starting. Attempting to connect to A: %s and C: %s", *aUrl, *cUrl)

	if err := bridge.run(ctx); err != nil {
		log.Fatalf("Bridge stopped due to error: %v", err)
	}

	log.Println("Bridge shut down gracefully.")
}
