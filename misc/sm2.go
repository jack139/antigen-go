package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"encoding/base64"
	"encoding/hex"
	"github.com/tjfoc/gmsm/sm2"
)


var (
	// 用户密钥
	privBase64 = string("vK3iPBFMwKvXfS6QG3s0fKNPjGnLy90VI+PI0kzQ3o0=")

	privKey *sm2.PrivateKey
	pubKey *sm2.PublicKey
)

func init(){
	// base64 恢复私钥
	privKey, _ = restoreKey(privBase64)
	//fmt.Printf("D: %x\nX: %x\nY: %x\n", priv.D, priv.PublicKey.X, priv.PublicKey.Y)

	// 公钥
	pubKey = &privKey.PublicKey
}

// 从 base64私钥 恢复密钥对
func restoreKey(privStr string) (*sm2.PrivateKey, error) {
	priv, err  := base64.StdEncoding.DecodeString(privStr)
	if err!=nil {
		return nil, err
	}

	curve := sm2.P256Sm2()
	key := new(sm2.PrivateKey)
	key.PublicKey.Curve = curve
	key.D = new(big.Int).SetBytes(priv)
	key.PublicKey.X, key.PublicKey.Y = curve.ScalarBaseMult(priv)
	return key, nil
}

// SM2签名
func SM2Sign(data []byte) ([]byte, error) {
	// 签名(内部已经做了sm3摘要)
	R, S, err := sm2.Sm2Sign(privKey, data, nil, rand.Reader) 
	if err!=nil {
		return nil, err
	}
	//fmt.Printf("R: %x\nS: %x\n", R, S)

	sign := R.Bytes()
	sign = append(sign, S.Bytes()...)

	return sign, nil
}

// SM2签名，返回base64编码
func SM2SignBase64(data []byte) (string, error) {
	sign, err := SM2Sign(data)
	if err!=nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(sign), nil
}

// SM2验签
func SM2Verify(data []byte, sign []byte) bool {
	if len(sign)<64 {
		return false
	}

	R := new(big.Int).SetBytes(sign[:32]) 
	S := new(big.Int).SetBytes(sign[32:])
	//fmt.Printf("R: %x\nS: %x\n", R, S)

	// 验签
	return sm2.Sm2Verify(pubKey, data, nil, R, S)
}

// SM2验签，使用base64编码
func SM2VerifyBase64(data []byte, signBase64 string) bool {
	sign, _  := base64.StdEncoding.DecodeString(signBase64)
	return SM2Verify(data, sign)
}


func main() {
	data := []byte("message digest")

	sign, _ := SM2Sign(data)
	fmt.Printf("sign: %x\n", sign)

	ok := SM2Verify(data, sign)
	if ok != true {
		log.Printf("SM2Verify() Verify error\n")
	} else {
		log.Printf("SM2Verify() Verify ok\n")
	}

	signBase64, _ := SM2SignBase64(data)
	ok = SM2VerifyBase64(data, signBase64)
	if ok != true {
		log.Printf("SM2VerifyBase64() Verify error\n")
	} else {
		log.Printf("SM2VerifyBase64() Verify ok\n")
	}


	// python 生成的签名
	//signStr2 := "1a0a614198d2705c6248285737d01b0e4d66e285ec77fab91d8c9da0bd82a3e27e802a14e5322d75eaee7fbeecd8a93d71f94692577f1f9b89f0f0f63acfbefa"
	signStr2 := "6461b7f242019bf4ebb11f47ef43af9e25611918e04675ac30e94dd7424c5d59ee003fefbc2c4a7b65d8d1bbd5b6e60b05179c7cf6e5a0fbab630fc951106204"

	fmt.Println("python signed: ", signStr2)

	// 验签
	sign, _ = hex.DecodeString(signStr2)
	ok = SM2Verify(data, sign)
	if ok != true {
		log.Printf("SM2Verify() Verify error\n")
	} else {
		log.Printf("SM2Verify() Verify ok\n")
	}

	signBase64 = base64.StdEncoding.EncodeToString(sign)
	ok = SM2VerifyBase64(data, signBase64)
	if ok != true {
		log.Printf("SM2VerifyBase64() Verify error\n")
	} else {
		log.Printf("SM2VerifyBase64() Verify ok\n")
	}

}