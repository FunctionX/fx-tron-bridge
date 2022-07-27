package utils

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	sdkcrypto "github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func DecryptFxPrivateKey(fxKeyValue, fxPwdValue string) (*secp256k1.PrivKey, error) {
	isFile, err := PathExists(fxKeyValue)
	if err != nil {
		return nil, err
	}
	if isFile {
		keyValueBytes, err := ioutil.ReadFile(fxKeyValue)
		if err != nil {
			return nil, err
		}
		existsPwdFile, err := PathExists(fxPwdValue)
		if err != nil {
			return nil, err
		}
		var pwdBytes []byte
		if existsPwdFile {
			pwdBytes, err = ioutil.ReadFile(fxPwdValue)
			if err != nil {
				return nil, err
			}
		}
		privateKey, _, err := sdkcrypto.UnarmorDecryptPrivKey(string(keyValueBytes), string(pwdBytes))
		if err != nil {
			return nil, err
		}
		orcPrivKey, ok := privateKey.(*secp256k1.PrivKey)
		if !ok {
			return nil, fmt.Errorf("privateKey not secp256k1 privateKey:%v", privateKey)
		}
		return orcPrivKey, nil
	} else if len(fxKeyValue) == 64 || (len(fxKeyValue) == 66 && strings.HasPrefix(fxKeyValue, "0x")) {
		keyBytes, err := hexutil.Decode(fxKeyValue)
		if err != nil {
			return nil, err
		}
		return &secp256k1.PrivKey{Key: keyBytes}, nil
	}
	return nil, fmt.Errorf("invalid private key")
}

func DecryptEthPrivateKey(tronKeyValue, tronPwdValue string) (*ecdsa.PrivateKey, error) {
	var tronKey *keystore.Key
	isFile, err := PathExists(tronKeyValue)
	if err != nil {
		return nil, err
	}
	if isFile {
		tronKeyValueBytes, err := ioutil.ReadFile(tronKeyValue)
		if err != nil {
			return nil, err
		}
		existsPwdFile, err := PathExists(tronPwdValue)
		if err != nil {
			return nil, err
		}
		var tronPwdBytes []byte
		if existsPwdFile {
			tronPwdBytes, err = ioutil.ReadFile(tronPwdValue)
			if err != nil {
				return nil, err
			}
			tronKey, err = keystore.DecryptKey(tronKeyValueBytes, string(tronPwdBytes))
			if err != nil {
				return nil, err
			}
		} else {
			tronKey, err = keystore.DecryptKey(tronKeyValueBytes, tronPwdValue)
			if err != nil {
				return nil, err
			}
		}
		return tronKey.PrivateKey, nil
	} else if len(tronKeyValue) == 64 || (len(tronKeyValue) == 66 && strings.HasPrefix(tronKeyValue, "0x")) {
		return crypto.HexToECDSA(tronKeyValue)
	}
	return nil, fmt.Errorf("invalid private key")
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
