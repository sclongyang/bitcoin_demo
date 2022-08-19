package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"math/big"
)

type Wallet struct {
	KeyPairs map[string] /*Addr*/ KeyPair
}

type KeyPair struct {
	PrivKey *ecdsa.PrivateKey
	PubKey  []byte
}

var WalletLocal Wallet = Wallet{
	KeyPairs: make(map[string]KeyPair, 5),
}

func (w *Wallet) ListAllAddr() {
	for addr, _ := range w.KeyPairs {
		fmt.Println("Addr: ", addr)
	}
}

func (w *Wallet) Send(fromAddr, targetAddr string, amount uint64, chain *BlockChain, minerAddr string) error {
	tx, err := NewTx(fromAddr, targetAddr, amount, chain, w)
	if err != nil {
		return err
	}
	return chain.AddBlock([]*Transaction{tx}, minerAddr, "矿工挖矿")
}

func GenKeyPair() error {
	p256 := elliptic.P256()
	privateKey, err := ecdsa.GenerateKey(p256, rand.Reader)
	if err != nil {
		return err
	}

	pubKeyRaw := privateKey.PublicKey
	pubKey := append(pubKeyRaw.X.Bytes(), pubKeyRaw.Y.Bytes()...) //简单起见使用非压缩的公钥
	addr := PubKey2Addr(pubKey)

	WalletLocal.KeyPairs[addr] = KeyPair{
		PrivKey: privateKey,
		PubKey:  pubKey,
	}

	fmt.Println("new Addr:", addr)
	return nil
}

func CheckAddr(addr string) bool {
	decode := base58.Decode(addr)
	if len(decode) != 25 {
		return false
	}
	payload := decode[:len(decode)-4]
	sum := checkSum(payload)
	return bytes.Equal(sum, decode[len(decode)-4:])
}

func PubKeyBytes2PubKey(pubkey []byte) (*ecdsa.PublicKey, error) {
	if len(pubkey) != 512 {
		return nil, fmt.Errorf("pubkey len is not 512")
	}
	var x, y big.Int
	x.SetBytes(pubkey[:len(pubkey)/2])
	y.SetBytes(pubkey[len(pubkey)/2:])
	p256 := elliptic.P256()
	return &ecdsa.PublicKey{
		Curve: p256,
		X:     &x,
		Y:     &y,
	}, nil
}

func Addr2PubKeyHash(addr string) (pubKeyHash []byte, err error) {
	if !CheckAddr(addr) {
		err = fmt.Errorf("地址格式不对:", addr)
		return
	}
	decode := base58.Decode(addr)
	pubKeyHash = decode[1 : len(decode)-4]
	return
}

func PubKeyHash2Addr(pubKeyHash []byte) string {
	payload := append([]byte{0x00}, pubKeyHash...)
	checksum := checkSum(payload)
	return base58.Encode(append(payload, checksum...))
}

func PubKey2Addr(pubKey []byte) string {
	pubKeyHash := Hash160(pubKey)
	return PubKeyHash2Addr(pubKeyHash)
}

func checkSum(payload []byte) []byte {
	dHash := Hash256(payload)
	return dHash[:4]
}
