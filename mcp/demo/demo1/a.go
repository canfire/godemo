package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const (
	targetHost = "127.0.0.1" // 目标 IP
	targetPort = "9222"      // 目标端口
)

// WebSocket 配置
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	http.HandleFunc("/", handleWS)
	log.Println("WebSocket 转发代理已启动: ws://localhost:8080/<任意路径>")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleWS(w http.ResponseWriter, r *http.Request) {
	// 拿到客户端原始路径
	path := r.URL.Path
	log.Println("收到客户端路径:", path)

	// 升级客户端连接
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade 失败:", err)
		return
	}
	defer clientConn.Close()

	// 拼接目标地址
	targetURL := "ws://" + targetHost + ":" + targetPort + path
	log.Println("转发到目标地址:", targetURL)

	// 连接目标 WebSocket
	targetConn, _, err := websocket.DefaultDialer.Dial(targetURL, nil)
	if err != nil {
		log.Println("连接目标 WS 失败:", err)
		return
	}
	defer targetConn.Close()

	// 客户端 → 目标 WS
	go func() {
		for {
			mt, msg, err := clientConn.ReadMessage()
			if err != nil {
				log.Println("客户端断开:", err)
				targetConn.Close()
				return
			}
			err = targetConn.WriteMessage(mt, msg)
			if err != nil {
				log.Println("转发到目标 WS 失败:", err)
				return
			}
		}
	}()

	// 目标 WS → 客户端
	for {
		mt, msg, err := targetConn.ReadMessage()
		if err != nil {
			log.Println("目标 WS 断开:", err)
			clientConn.Close()
			return
		}
		err = clientConn.WriteMessage(mt, msg)
		if err != nil {
			log.Println("转发到客户端失败:", err)
			return
		}
	}
}
