package transport

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io"

	"github.com/pion/webrtc/v3"
)

func EncodeSDP(obj *webrtc.SessionDescription) string {
	b, _ := json.Marshal(obj)
	var buf bytes.Buffer
	gz, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	gz.Write(b)
	gz.Close()
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func DecodeSDP(s string) webrtc.SessionDescription {
	compressed, _ := base64.StdEncoding.DecodeString(s)
	gz, _ := gzip.NewReader(bytes.NewReader(compressed))
	decompressed, _ := io.ReadAll(gz)
	gz.Close()

	var sdp webrtc.SessionDescription
	json.Unmarshal(decompressed, &sdp)
	return sdp
}
