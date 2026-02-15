package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/pion/webrtc/v3"

	"XynerraP2P/internal/console"
	"XynerraP2P/internal/transport"
	"XynerraP2P/internal/webrtcpeer"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Run() {
	rl, err := readline.New("> ")
	console.Check(err)
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

	peer := webrtcpeer.NewPeer(username, safePrint)
	defer peer.PC.Close()

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

		d, _ := peer.PC.CreateDataChannel("chat", &webrtc.DataChannelInit{
			Ordered:        &ordered,
			MaxRetransmits: &maxRetransmits,
		})
		peer.SetupDataChannel(d)

		offer, _ := peer.PC.CreateOffer(nil)
		gather := webrtc.GatheringCompletePromise(peer.PC)
		peer.PC.SetLocalDescription(offer)
		<-gather

		fmt.Println("\n===== OFFER =====")
		fmt.Println(transport.EncodeSDP(peer.PC.LocalDescription()))
		fmt.Println("=================\n")
		fmt.Println("Paste ANSWER:")

		answer := readLongInput()
		console.Clear()
		peer.PC.SetRemoteDescription(transport.DecodeSDP(answer))

	} else {
		peer.PC.OnDataChannel(func(d *webrtc.DataChannel) {
			peer.SetupDataChannel(d)
		})

		fmt.Println("Paste OFFER:")
		offer := readLongInput()
		peer.PC.SetRemoteDescription(transport.DecodeSDP(offer))

		answer, _ := peer.PC.CreateAnswer(nil)
		gather := webrtc.GatheringCompletePromise(peer.PC)
		peer.PC.SetLocalDescription(answer)
		<-gather

		fmt.Println("\n===== ANSWER =====")
		fmt.Println(transport.EncodeSDP(peer.PC.LocalDescription()))
		fmt.Println("==================")
		fmt.Println("")
	}

	rl, _ = readline.New("> ")
	defer rl.Close()

	for {
		text, err := rl.Readline()
		if err != nil {
			break
		}

		if peer.DC != nil && strings.TrimSpace(text) != "" {
			b := []byte(`{"c":"` + text + `"}`)
			peer.DC.Send(b)
		}
	}
}
