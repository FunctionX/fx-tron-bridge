package bridge

import (
	"testing"
)

func TestSigner(t *testing.T) {
	testClient, err := GetTestClient()
	if err != nil {
		t.Fatal(err)
	}
	singer, err := NewSinger(testClient, "FX")
	if err != nil {
		t.Fatal(err)
	}
	if err = singer.signer(); err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestSetFxKeyBalanceMetrics(t *testing.T) {
	testClient, err := GetTestClient()
	if err != nil {
		t.Fatal(err)
	}
	singer, err := NewSinger(testClient, "FX")
	if err != nil {
		t.Fatal(err)
	}
	err = singer.setFxKeyBalanceMetrics()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
