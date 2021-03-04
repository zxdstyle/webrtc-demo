<template>
  <div class="server">
    <video id="localVideo" autoplay playsinline muted></video>
  </div>
</template>

<script>
import "webrtc-adapter"
export default {
  name: "Home",
  setup() {
    let ws = new WebSocket("ws://127.0.0.1:8199/client")
    // 创建PeerConnection实例 (参数为null则忽略 iceserver，仅局域网下通讯)
    let peer = new RTCPeerConnection(null)

    ws.onmessage = event => {
      const { type, sdp, iceCandidate } = JSON.parse(event.data)
      if (type === "offer") {
        answer(new RTCSessionDescription({ type, sdp }))
      } else if (type === "offer_ice") {
        peer.addIceCandidate(iceCandidate)
      }
    }

    const answer = async sdp => {
      await peer.setRemoteDescription(sdp)

      const answer = await peer.createAnswer()

      ws.send(JSON.stringify(answer))

      await peer.setLocalDescription(answer) 
    }

    peer.ontrack = e => {
      if (e && e.streams) {
        console.log("收到对方音频/视频流数据...")
        let localVideo = document.getElementById("localVideo")
        localVideo.srcObject = e.streams[0]
        console.log(localVideo, e.streams)
      }
    }


    peer.onicecandidate = e => {ç
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


    return {  }
  }
}
</script>

<style scoped>

.server {

}

video {
  width: 80%;
  height: 80%;
  background: #000;
  position: absolute;
  left: 0;right: 0;
  top: 0;bottom: 0;
  margin: auto;
}


button {
  position: absolute;
  left: 0;right: 0;
  top: 0;bottom: 0;
  margin: auto;
  width: 130px;
  height: 40px;
  border-radius: 5px;
  padding: 10px 25px;
  font-weight: 500;
  background: #fff;
  cursor: pointer;
  transition: all 0.3s ease;
  display: inline-block;
  border: none;
  color: #000;
  outline: none;
}
</style>