package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/chzyer/readline"
	"github.com/pion/webrtc/v3"
)

type ChatMessage struct {
	Content string `json:"c"`
}

func main() {
	rl, err := readline.New("> ")
	check(err)
	defer rl.Close()

	rl.SetPrompt("Username > ")
	rl.Refresh()
	username, _ := rl.Readline()
	username = strings.TrimSpace(username)
	if username == "" {
		username = "anon"
	}
	rl.SetPrompt("> ")
	rl.Refresh()

	safePrint := func(s string) {
		rl.Write([]byte(s + "\n"))
	}

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}

	pc, err := webrtc.NewPeerConnection(config)
	check(err)
	defer pc.Close()

	var dc *webrtc.DataChannel
	remoteName := "peer"

	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		safePrint("[Xynerra] - ICE State: " + state.String())
	})

	onMessage := func(msg webrtc.DataChannelMessage) {
		var m ChatMessage
		check(json.Unmarshal(msg.Data, &m))

		if strings.HasPrefix(m.Content, "__hello__:") {
			remoteName = strings.TrimPrefix(m.Content, "__hello__:")
			ClearConsole()
			safePrint("[Xynerra] - Connected to: " + remoteName)
			fmt.Println()
			return
		}

		safePrint("[" + remoteName + "]: " + m.Content)
	}

	setupDC := func(d *webrtc.DataChannel) {
		d.OnOpen(func() {
			hello := ChatMessage{Content: "__hello__:" + username}
			b, _ := json.Marshal(hello)
			d.Send(b)
			safePrint("[Xynerra] - Connected.")
		})
		d.OnMessage(onMessage)
	}

	safePrint("1: Offer")
	safePrint("2: Answer")

	choice, _ := rl.Readline()

	readLongInput := func() string {
		rl.Close()
		scanner := bufio.NewScanner(os.Stdin)
		buf := make([]byte, 1024*1024)
		scanner.Buffer(buf, 1024*1024)
		scanner.Scan()
		return strings.TrimSpace(scanner.Text())
	}

	if choice == "1" {
		ordered := true
		maxRetransmits := uint16(0)

		d, _ := pc.CreateDataChannel("chat", &webrtc.DataChannelInit{
			Ordered:        &ordered,
			MaxRetransmits: &maxRetransmits,
		})
		dc = d
		setupDC(d)

		offer, _ := pc.CreateOffer(nil)
		gather := webrtc.GatheringCompletePromise(pc)
		pc.SetLocalDescription(offer)
		<-gather

		fmt.Println("\n===== OFFER =====")
		fmt.Println(encodeSDP(pc.LocalDescription()))
		fmt.Println("=================\n")
		fmt.Println("Paste ANSWER:")

		answer := readLongInput()
		ClearConsole()
		pc.SetRemoteDescription(decodeSDP(answer))

	} else {
		pc.OnDataChannel(func(d *webrtc.DataChannel) {
			dc = d
			setupDC(d)
		})

		fmt.Println("Paste OFFER:")
		offer := readLongInput()
		pc.SetRemoteDescription(decodeSDP(offer))

		answer, _ := pc.CreateAnswer(nil)
		gather := webrtc.GatheringCompletePromise(pc)
		pc.SetLocalDescription(answer)
		<-gather

		fmt.Println("\n===== ANSWER =====")
		fmt.Println(encodeSDP(pc.LocalDescription()))
		fmt.Println("==================\n")
	}

	rl, _ = readline.New("> ")
	defer rl.Close()

	for {
		text, err := rl.Readline()
		if err != nil {
			break
		}

		if dc != nil && strings.TrimSpace(text) != "" {
			b, _ := json.Marshal(ChatMessage{Content: text})
			dc.Send(b)
		}
	}
}

func encodeSDP(obj *webrtc.SessionDescription) string {
	b, _ := json.Marshal(obj)
	var buf bytes.Buffer
	gz, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	gz.Write(b)
	gz.Close()
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func decodeSDP(s string) webrtc.SessionDescription {
	compressed, _ := base64.StdEncoding.DecodeString(s)
	gz, _ := gzip.NewReader(bytes.NewReader(compressed))
	decompressed, _ := io.ReadAll(gz)
	gz.Close()

	var sdp webrtc.SessionDescription
	json.Unmarshal(decompressed, &sdp)
	return sdp
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func ClearConsole() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default: // linux, darwin, or smth
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
}
