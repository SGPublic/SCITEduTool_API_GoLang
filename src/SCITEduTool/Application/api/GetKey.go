package api

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"
)

type KeyResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Hash    string `json:"hash"`
	Key     string `json:"key"`
}

func GetKey(w http.ResponseWriter, _ *http.Request) {
	keyResult := KeyResult{
		Code:    200,
		Message: "success.",
		Hash:    GetRandomString(8),
		Key: "-----BEGIN PUBLIC KEY-----\n" +
			"MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCmBCWNtxeofYkH1e9GXKgszj4E\n" +
			"cJojNvlesPDM201q+fiVf2X4SWPNjdduRS19dq9Koq4Dz0ul3xV6E3ydCHl88qSa\n" +
			"94fDGZa24UueYVYE0ytYuJcOu164GlIfu48Rir0NXA2BfoQxMcSpMmLJt20rSg+E\n" +
			"oP24zaj3ti78b1zJEwIDAQAB\n" +
			"-----END PUBLIC KEY-----",
	}
	result, _ := json.Marshal(keyResult)
	_, _ = w.Write(result)
}

func GetRandomString(l int) string {
	str := "0123456789abcdef0123456789ghijkl0123456789mnopqr0123456789stuvwx0123456789yz"
	bytes := []byte(str)
	var result []byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
