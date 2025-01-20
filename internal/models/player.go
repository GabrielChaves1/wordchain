package models

import "github.com/gorilla/websocket"

type Player struct {
	Name       string
	Connection *websocket.Conn
	IsReady    bool
}
