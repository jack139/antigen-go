package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"encoding/base64"
	"github.com/tjfoc/gmsm/sm2"
)


var (
	// 用户密钥
	daPriStr = string("vK3iPBFMwKvXfS6QG3s0fKNPjGnLy90VI+PI0kzQ3o0=")
)


// 从 base64私钥 恢复密钥对
func restoreKey(privStr string) *sm2.PrivateKey {
	priv, _  := base64.StdEncoding.DecodeString(privStr)
	fmt.Printf("priv %d %v\n", len(priv), priv)

	curve := sm2.P256Sm2()
	key := new(sm2.PrivateKey)
	key.PublicKey.Curve = curve
	key.D = new(big.Int).SetBytes(priv)
	key.PublicKey.X, key.PublicKey.Y = curve.ScalarBaseMult(priv)
	return key
}


func main() {
	data := []byte("message digest")

	// base64 恢复私钥
	priv := restoreKey(daPriStr)
	fmt.Printf("D: %x\nX: %x\nY: %x\n", priv.D, priv.PublicKey.X, priv.PublicKey.Y)

	// 公钥
	pubKey := priv.PublicKey

	// 验证sm3摘要结果
	sm3DigestData, _ := pubKey.Sm3Digest(data, nil)
	fmt.Printf("Sm3Digest: %x\n", sm3DigestData)

	sign, err := priv.Sign(rand.Reader, data, nil) // 签名(内部已经做了sm3摘要)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("sign: %x\n", sign)

	R1, S1, _ := sm2.SignDataToSignDigit(sign) // 拆分出 R S
	fmt.Printf("R1: %x\nS1: %x\n", R1, S1)

	sign1, _ := sm2.SignDigitToSignData(R1, S1)  // 重新合并 R S
	fmt.Printf("sign1: %x\n", sign1)

	ok := pubKey.Verify(data, sign1) // 密钥验证
	if ok != true {
		log.Printf("Verify error\n")
	} else {
		log.Printf("Verify ok\n")
	}

	// python 生成的签名
	signStr2 := "1a0a614198d2705c6248285737d01b0e4d66e285ec77fab91d8c9da0bd82a3e27e802a14e5322d75eaee7fbeecd8a93d71f94692577f1f9b89f0f0f63acfbefa"
	//signStr2 := "6461b7f242019bf4ebb11f47ef43af9e25611918e04675ac30e94dd7424c5d59ee003fefbc2c4a7b65d8d1bbd5b6e60b05179c7cf6e5a0fbab630fc951106204"
	R2, _ := new(big.Int).SetString(signStr2[:64], 16) // 拆分 R S
	S2, _ := new(big.Int).SetString(signStr2[64:], 16)
	fmt.Printf("R2: %x\nS2: %x\n", R2, S2)

	sign2, _ := sm2.SignDigitToSignData(R2, S2) // 合并 R S
	fmt.Printf("sign2: %x\n", sign2)

	ok = pubKey.Verify(data, sign2) // 密钥验证
	if ok != true {
		log.Printf("Verify error\n")
	} else {
		log.Printf("Verify ok\n")
	}

}