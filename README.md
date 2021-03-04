### webrtc 通话流程

1. ClientA 和 ClientB 都链接上信令服务器（websocket）
2. ClientA 获取本地视频流，并通过（websocket）发送会话描述信息（offer sdp）
3. ClientB 收到信令（offer sdp）后回复（answer sdp）
4. ClientA 接收到回复信令后开始链接，协商通信密钥，完成视频传输。

<!-- more -->

> 值得注意的是，webrtc 是客户端与客户端直接链接，不需要服务端参与。信令服务器仅在建立链接阶段交换数据。

### 基础环境

首先搭建一个基础环境方便开发，我使用的是尤大最新开发的vite。

```bash
yarn create @vitejs/app webrtc --template vue && cd webrtc && yarn && yarn dev
```

安装webrtc浏览器兼容适配器：

```bash
yarn add webrtc-adapter
```

一个小的demo就不引入不必要的扩展包了，自己动手实现一个简单的路由器。调整`App.vue` 文件：

```js
<script>
import { h, computed } from "vue"
import routes from "./routes"

export default {
  setup() {
    const currentRoute = computed(() => window.location.pathname)
    const currentComponent = computed(() => routes[currentRoute.value] || "")

    return () => h(currentComponent.value)
  }
}

</script>

```

新建 `src/routes.js` 文件：

```js
import Home from "./page/Server.vue";
import Client from "./page/Client.vue";

export default {
    '/': Home,
    '/client': Client
}
```



这样就可以分开编写客户端和服务端代码了。由于这只是个简单的 demo ，所以我这样使用，生成环境开发还是建议大家使用 `vue-router`。



### 正文

首先创建 `src/page/Server.vue`组件,引入 `webrtc`适配器：

```html
<template>
  <div class="server">
    <video id="localVideo" autoplay playsinline muted></video>
  </div>
</template>

<script>
import "webrtc-adapter"
export default {
  name: "Server"
}
</script>

<style scoped>
video {
  width: 80%;
  height: 80%;
  background: #000;
  position: absolute;
  left: 0;right: 0;
  top: 0;bottom: 0;
  margin: auto;
}
</style>
```



#### 1. 获取本地视频流

这一步既可以录制屏幕内容也可以读取用户摄像头。

```js
<template>
  <div class="server">
    <video id="localVideo" autoplay playsinline muted></video>
  </div>
</template>

<script>
import "webrtc-adapter"
import { onMounted } from "vue"

export default {
  name: "Home",
  setup() {
     const startLive = async () => {
       const localVideo = document.getElementById("localVideo")
       let stream = await navigator.mediaDevices.getDisplayMedia({ video: true, audio: true })
       localVideo.srcObject = stream
     }
     
     onMounted(() => startLive())

     return {  }
  }
}
</script>
```

> 这里需要注意的是一定要在dom节点加载完毕之后再执行，否则会获取不到video节点



#### 2. 信令服务器

想要与客户端建立链接，必须有信令服务器端介入，帮助客户端之间进行通讯。其实信令服务器原理很简单，不用做任何处理（业务逻辑除外），只需要把客户端发送的数据原样转发给目标客户端即可。我这里采用 GoLang 实现：

```go
package main

import (
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/glog"
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

```



#### 3. 建立链接

链接信令服务器并且创建 PeerConnection 实例 (参数为null，忽略 iceserver，仅局域网下通讯)；

```js
let ws = new WebSocket("ws://127.0.0.1:8199/server")
let peer = new RTCPeerConnection(null)

// 将媒体轨道添加到轨道集
stream.getTracks().forEach(track => {
  peer.addTrack(track, stream)
})
```



然后通过websocket发送会话描述信息（offer sdp）：

```js
const offer = await peer.createOffer()
await peer.setLocalDescription(offer)
ws.send(JSON.stringify(offer))
```



#### 4. 客户端接收

客户端同样链接信令服务器并创建PeerConnection实例。

```js
let ws = new WebSocket("ws://127.0.0.1:8199/client")
let peer = new RTCPeerConnection(null)
```

监听信令服务器消息：

```js
ws.onmessage = event => {
  const { type, sdp, iceCandidate } = JSON.parse(e.data)
  if (type === "offer") {
    answer(new RTCSessionDescription({ type, sdp }))
  } else if (type === "offer_ice") {
    peer.addIceCandidate(iceCandidate)
  }
}
```

收到服务器链接请求后，需要回复消息：

```js
const answer = async sdp => {
  await peer.setRemoteDescription(sdp)

  const answer = await peer.createAnswer()

  ws.send(JSON.stringify(answer))

  await peer.setLocalDescription(answer) 
}
```



#### 5. 服务端监听响应

客户端收到链接请求并回复，服务端同时也要监听客户端的回复：

```js
ws.onmessage = e => {
  const { type, sdp, iceCandidate } = JSON.parse(e.data)
  if (type === "answer") {
    peer.setRemoteDescription(new RTCSessionDescription({ type, sdp }))
  } else if (type === "answer_ice") {
    peer.addIceCandidate(iceCandidate)
  }
}
```



#### 6. 获取远程视频流

当调用 `setLocalDescription` 之后，RTC 链接就开始搜集候选人，而我们只需要监听事件，通过信令服务器传递候选人信息即可。

```js
// 服务端
peer.onicecandidate = e => {
  if (e.candidate) {
    console.log("搜集并发送候选人")
    ws.send(
      JSON.stringify({
        type: `offer_ice`,
        iceCandidate: e.candidate,
      })
    )
  } else {
    console.log("候选人收集完成！")
  }
}

// 客户端
peer.onicecandidate = e => {
  if (e.candidate) {
    console.log("搜集并发送候选人")
    ws.send(
      JSON.stringify({
        type: `answer_ice`,
        iceCandidate: e.candidate,
      })
    )
  } else {
    console.log("候选人收集完成！")
  }
}
```

  候选人搜集完成之后，服务端和客户端就正式勾搭上了，开始通信协商密钥，建立一条最优的链接方式。这个时候只需要监听 `ontrack` 事件获取视频流即可：

```js
peer.ontrack = e => {
  if (e && e.streams) {
    console.log("收到对方音频/视频流数据...")
    let localVideo = document.getElementById("localVideo")
    localVideo.srcObject = e.streams[0]
  }
}
```



### 结语

RTCPeerConnection 链接是双向的，这里我为了演示方便，仅做了单向视频传输，强行区分了服务端与客户端。客户端收到链接请求的同时也可以获取本地视频流传递到服务端，形成双向视频传输。同时为了方便演示信令服务器也是最简单的方式实现，客户端比需先链接才能进行通讯。
