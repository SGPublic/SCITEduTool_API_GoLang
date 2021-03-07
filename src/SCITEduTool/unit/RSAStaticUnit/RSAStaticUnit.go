package RSAStaticUnit

import (
	"SCITEduTool/unit/StdOutUnit"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
)

var GetPrivateKey func() []byte

func DecodePublicEncode(data string) (string, StdOutUnit.MessagedError) {
	priKey, err := x509.ParsePKCS1PrivateKey(GetPrivateKey())
	if err != nil {
		return "", StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	dataBase64, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	dataDecrypted, err := rsa.DecryptPKCS1v15(rand.Reader, priKey, dataBase64)
	if err != nil {
		return "", StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	} else {
		return string(dataDecrypted), StdOutUnit.GetEmptyErrorMessage()
	}
}
