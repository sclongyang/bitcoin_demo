package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"golang.org/x/crypto/ripemd160"
)

//double sha256
func Hash256(data []byte) []byte {
	first := sha256.Sum256(data)
	second := sha256.Sum256(first[:])
	return second[:]
}

//ripemd160(sha256(data))
func Hash160(data []byte) []byte {
	first := sha256.Sum256(data)
	ripeHash := ripemd160.New()
	ripeHash.Write(first[:])
	return ripeHash.Sum(nil)
}

func Convert2Bytes(num uint64) []byte {
	var buffer bytes.Buffer
	err := binary.Write(&buffer, binary.LittleEndian, num)
	if err != nil {
		fmt.Println("Convert2Bytes err:", num)
		return nil
	}
	return buffer.Bytes()
}
