package fxtronbridge

import (
	"bytes"
	"net/http"
	"testing"
)

func TestStartBridgePrometheus(t *testing.T) {
	StartBridgePrometheus()
	BlockHeightProm.Set(100)
	BlockIntervalProm.Set(101)
	MsgPendingLenProm.Inc()
	client := http.DefaultClient
	get, err := client.Get("http://127.0.0.1:9811")
	if err != nil {
		t.Fatal(err)
	}
	var buf = new(bytes.Buffer)
	_, err = buf.ReadFrom(get.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(buf.String())
}
