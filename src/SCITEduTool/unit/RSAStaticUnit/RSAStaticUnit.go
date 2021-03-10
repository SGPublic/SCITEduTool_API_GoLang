package RSAStaticUnit

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
	"strings"

	"SCITEduTool/unit/StdOutUnit"
)

var privateKey *rsa.PrivateKey

type PrivateKey struct {
	Content string `json:"content"`
}

func DecodePublicEncode(data string) (string, StdOutUnit.MessagedError) {
	dataBase64, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		StdOutUnit.Error.String("", "数据不是base64数据", err)
		return "", StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	dataDecrypted, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, dataBase64)
	if err != nil {
		StdOutUnit.Error.String("", "数据不是RSA加密数据", err)
		return "", StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	} else {
		return string(dataDecrypted), StdOutUnit.GetEmptyErrorMessage()
	}
}

func SetPrivateKey(keyConf PrivateKey) {
	if keyConf.Content == "" || strings.Contains(keyConf.Content, "//") {
		StdOutUnit.Assert.String("", "私钥数据为空或格式错误", nil)
		os.Exit(0)
	}
	key, err := base64.StdEncoding.DecodeString(keyConf.Content)
	if err != nil {
		StdOutUnit.Assert.String("", "私钥数据解密失败", err)
		os.Exit(0)
	}
	block, _ := pem.Decode(key)
	privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		StdOutUnit.Assert.String("", "私钥解析失败", err)
		os.Exit(0)
	}
	StdOutUnit.Verbose.String("", "RSA配置成功")
}
