package main

import (
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/glog"
	"github.com/gorilla/websocket"
)

var (
	Server *ghttp.WebSocket
	Client *ghttp.WebSocket
)

func main()  {
	s := g.Server()

	s.BindHandler("/server", func(r *ghttp.Request) {
		ws, err := r.WebSocket()
		if err != nil {
			glog.Error(err)
			r.Exit()
		}

		Server = ws

		if Client == nil {
			errMsg, _ := gjson.Encode(map[string]interface{} {
				"error": 400,
				"message": "客户端未链接",
			})
			Server.WriteMessage(websocket.TextMessage,errMsg)
			Server.Close()
		} else {
			msg, _ := gjson.Encode(map[string]interface{} {
				"error": 200,
				"message": "客户端已链接",
			})
			Server.WriteMessage(websocket.TextMessage,msg)
		}

		for {
			msgType, msg, err := ws.ReadMessage()
			if err != nil {
				return
			}

			if err = Client.WriteMessage(msgType, msg); err != nil {
				return
			}
		}
	})

	s.BindHandler("/client", func(r *ghttp.Request) {
		ws, err := r.WebSocket()
		if err != nil {
			glog.Error(err)
			r.Exit()
		}

		Client = ws

		for {
			msgType, msg, err := ws.ReadMessage()
			if err != nil {
				return
			}
			if err = Server.WriteMessage(msgType, msg); err != nil {
				return
			}
		}
	})

	s.SetPort(8199)
	s.Run()
}
