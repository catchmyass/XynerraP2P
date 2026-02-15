package webrtcpeer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pion/webrtc/v3"

	"XynerraP2P/internal/console"
	"XynerraP2P/internal/model"
)

type Peer struct {
	PC         *webrtc.PeerConnection
	DC         *webrtc.DataChannel
	Username   string
	RemoteName string
	SafePrint  func(string)
}

func NewPeer(username string, safePrint func(string)) *Peer {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}

	pc, err := webrtc.NewPeerConnection(config)
	console.Check(err)

	p := &Peer{
		PC:         pc,
		Username:   username,
		RemoteName: "peer",
		SafePrint:  safePrint,
	}

	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		p.SafePrint("[Xynerra] - ICE State: " + state.String())
	})

	return p
}

func (p *Peer) SetupDataChannel(d *webrtc.DataChannel) {
	p.DC = d

	d.OnOpen(func() {
		hello := model.ChatMessage{Content: "__hello__:" + p.Username}
		b, _ := json.Marshal(hello)
		d.Send(b)
		p.SafePrint("[Xynerra] - Connected.")
	})

	d.OnMessage(func(msg webrtc.DataChannelMessage) {
		var m model.ChatMessage
		console.Check(json.Unmarshal(msg.Data, &m))

		if strings.HasPrefix(m.Content, "__hello__:") {
			p.RemoteName = strings.TrimPrefix(m.Content, "__hello__:")
			console.Clear()
			p.SafePrint("[Xynerra] - Connected to: " + p.RemoteName)
			fmt.Println()
			return
		}

		p.SafePrint("[" + p.RemoteName + "]: " + m.Content)
	})
}
