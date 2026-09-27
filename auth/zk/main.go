package main

import (
	"fmt"
	"math/big"

	"github.com/Crows-Storm/Axis/auth/zk/auth"
	"github.com/Crows-Storm/Axis/auth/zk/crypto"
	"github.com/Crows-Storm/Axis/auth/zk/schnorr"
)

func main() {
	store := auth.NewUserStore()

	// mock a username and password and salt
	username := "alice"
	password := "s3cr3t-passw0rd"
	salt := []byte("random-salt-16by")

	x := schnorr.DeriveKeyFromPassword([]byte(password), salt)
	kp := schnorr.GenerateKeyPair(x)

	if err := store.Register(username, salt, kp.PublicX.Bytes(), kp.PublicY.Bytes()); err != nil {
		panic(err)
	}
	fmt.Println("[注册] 用户 alice 注册成功，服务器只存公钥")

	nonce, err := store.IssueChallenge(username)
	if err != nil {
		panic(err)
	}
	ctx := append([]byte("login:"+username+":"), nonce...)
	proof, err := schnorr.Prove(kp, ctx)
	if err != nil {
		panic(err)
	}
	err = store.VerifyLogin(username, proof)
	fmt.Printf("[登录] 验证结果: %v（正确密码）\n", err == nil)

	xWrong := schnorr.DeriveKeyFromPassword([]byte("wrong"), salt)
	kpWrong := schnorr.GenerateKeyPair(xWrong)
	nonce2, err := store.IssueChallenge(username)
	if err != nil {
		panic(err)
	}
	ctx2 := append([]byte("login:"+username+":"), nonce2...)
	proofWrong, err := schnorr.Prove(kpWrong, ctx2)
	if err != nil {
		panic(err)
	}
	err = store.VerifyLogin(username, proofWrong)
	fmt.Printf("[登录] 验证结果: %v（错误密码，应为 false）\n", err == nil)

	userPermission := big.NewInt(850)
	threshold := big.NewInt(500)

	r, err := crypto.RandomScalar()
	if err != nil {
		panic(err)
	}
	permProof, err := auth.ProvePermissionAbove(userPermission, threshold, r)
	if err != nil {
		panic(err)
	}
	ok := auth.VerifyPermissionAbove(permProof, threshold)
	fmt.Printf("[权限] 证明 v >= %v 结果: %v（实际 v=850，服务端不知道）\n", threshold, ok)

	lowPerm := big.NewInt(100)
	_, err = auth.ProvePermissionAbove(lowPerm, threshold, r)
	fmt.Printf("[权限] 用户 v=100 尝试证明 v>=500: %v（应失败）\n", err)
}

func ClientLogin(store *auth.UserStore, username, password string) error {
	nonce, err := store.IssueChallenge(username)
	if err != nil {
		return err
	}
	salt, err := store.Salt(username)
	if err != nil {
		return err
	}
	x := schnorr.DeriveKeyFromPassword([]byte(password), salt)
	kp := schnorr.GenerateKeyPair(x)
	ctx := append([]byte("login:"+username+":"), nonce...)
	proof, err := schnorr.Prove(kp, ctx)
	if err != nil {
		return err
	}
	return store.VerifyLogin(username, proof)
}
