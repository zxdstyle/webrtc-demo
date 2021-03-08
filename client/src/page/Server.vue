<template>
  <div class="server">
    <video id="localVideo" autoplay playsinline muted></video>
  </div>
</template>

<script>
import "webrtc-adapter"
import {onMounted} from "vue"

export default {
  name: "Home",
  setup() {
    let ws = new WebSocket("ws://127.0.0.1:8199/server")
    // 创建PeerConnection实例 (参数为null则忽略 iceserver，仅局域网下通讯)
    let peer = new RTCPeerConnection(null)

    ws.onerror = () => alert("信令服务链接异常")

    ws.onmessage = e => {
      let res = JSON.parse(e.data)
      if (res.error === 400) {
          return alert(res.message)
      } else if (res.error === 200) {
        startLive()
      }


      const { type, sdp, iceCandidate } = JSON.parse(e.data)
      if (type === "answer") {
        peer.setRemoteDescription(new RTCSessionDescription({ type, sdp }))
      } else if (type === "answer_ice") {
        peer.addIceCandidate(iceCandidate)
      }
    }

     const startLive = async () => {
       const localVideo = document.getElementById("localVideo")
       let stream = await navigator.mediaDevices.getDisplayMedia({ video: true, audio: true })
       localVideo.srcObject = stream

       // 将媒体轨道添加到轨道集
       stream.getTracks().forEach(track => {
         peer.addTrack(track, stream)
       })

       const offer = await peer.createOffer()
       await peer.setLocalDescription(offer)
       ws.send(JSON.stringify(offer))

     }

    peer.onicecandidate = e => {
      console.log(e)
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