package bridge

import (
	"github.com/functionx/fx-tron-bridge/client"
	"testing"
)

func TestGetLastBlockNumber(t *testing.T) {
	tronClient, err := client.NewTronGrpcClient("http://127.0.0.1:50051")
	if err != nil {
		t.Fatal(err)
	}
	latestBlockNumber, err := getLastBlockNumber("TVSMxNVuhzHTCvcnPzFmyAn2B2iDQjdgQh", tronClient)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("latestBlockNumber:", latestBlockNumber)
}

func TestHandleEvent(t *testing.T) {
	testClient, err := GetTestClient()
	if err != nil {
		t.Fatal(err)
	}
	oracle, err := NewOracle(testClient, 19512901-1)
	if err != nil {
		t.Fatal(err)
	}

	err = oracle.HandleEvent()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSaveLastBlockNumber(t *testing.T) {
	err := saveLastBlockNumber(900)
	if err != nil {
		t.Fatal(err)
	}

	lastBlockNumber, err := readLastBlockNumber()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("lastBlockNumber:", lastBlockNumber)
}

func TestReadLastBlockNumber(t *testing.T) {
	lastBlockNumber, err := readLastBlockNumber()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("lastBlockNumber:", lastBlockNumber)
}
