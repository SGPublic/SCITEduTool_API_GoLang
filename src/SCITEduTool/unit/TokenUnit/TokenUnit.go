package TokenUnit

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"SCITEduTool/manager/SessionManager"
	"SCITEduTool/unit/RSAStaticUnit"
	"SCITEduTool/unit/StdOutUnit"
)

type TokenConfig struct {
	TokenKey       string `json:"token_key"`
	TokenSecret    string `json:"token_secret"`
	AccessExpired  string `json:"access_expired"`
	RefreshExpired string `json:"refresh_expired"`
}

var tokenKey string
var tokenSecret string
var access int64
var refresh int64

type Token struct {
	AccessToken  string
	RefreshToken string
}

type TokenCheckResult struct {
	Username         string
	AccessEffective  bool
	RefreshEffective bool
}

type TokenBody struct {
	Password string
	Time     int64
	Type     string
	Key      string
}

func InitKey(tokenConf TokenConfig) {
	var err error
	if tokenConf.TokenKey == "" || tokenConf.TokenSecret == "" ||
		strings.Contains(tokenConf.TokenKey, "//") ||
		strings.Contains(tokenConf.TokenSecret, "//") {
		StdOutUnit.Assert("", "token_key或token_secret为空或格式不正确", nil)
		os.Exit(0)
	}
	tokenKey = tokenConf.TokenKey
	tokenSecret = tokenConf.TokenSecret
	access, err = strconv.ParseInt(tokenConf.AccessExpired, 10, 64)
	if err != nil {
		StdOutUnit.Warn("", "access_token过期时间解析失败，将使用默认值", err)
		access = 2592000
	}
	refresh, err = strconv.ParseInt(tokenConf.RefreshExpired, 10, 64)
	if err != nil {
		StdOutUnit.Warn("", "refresh_token过期时间解析失败，将使用默认值", err)
		refresh = 124416000
	} else {
		StdOutUnit.Verbose("", "Token配置成功")
	}
}

func Build(username string, password string) (Token, StdOutUnit.MessagedError) {
	token := Token{
		AccessToken:  "",
		RefreshToken: "",
	}
	timeNow := time.Now().Unix()
	headerPre := username + "&" + strconv.FormatInt(timeNow, 10)
	for len(headerPre)%3 != 0 {
		headerPre += "%"
	}
	header := base64.StdEncoding.EncodeToString([]byte(headerPre))
	token.AccessToken += header + "."

	passwordPre := password
	//IF DEBUG
	var err StdOutUnit.MessagedError
	passwordPre, err = RSAStaticUnit.DecodePublicEncode(passwordPre)
	if err.HasInfo {
		return Token{}, err
	}
	if len(passwordPre) <= 8 {
		return Token{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	passwordPre = passwordPre[8:]
	//ENDIF

	accessBodyPre, _ := json.Marshal(TokenBody{
		Password: passwordPre,
		Time:     timeNow,
		Type:     "access",
		Key:      tokenKey,
	})
	refreshBodyPre, _ := json.Marshal(TokenBody{
		Password: passwordPre,
		Time:     timeNow,
		Type:     "refresh",
		Key:      tokenKey,
	})
	token.AccessToken = getMD5(accessBodyPre) + "." + token.AccessToken
	token.RefreshToken = getMD5(refreshBodyPre) + "."

	accessFooterPre := token.AccessToken + tokenSecret
	refreshFooterPre := token.RefreshToken + header + "." + tokenSecret

	token.AccessToken += getMD5([]byte(accessFooterPre))
	token.RefreshToken += getFullMD5([]byte(refreshFooterPre))
	return token, StdOutUnit.GetEmptyErrorMessage()
}

func Check(token Token) (string, StdOutUnit.MessagedError) {
	username := ""
	if token.AccessToken == "" {
		if token.RefreshToken != "" {
			StdOutUnit.Info("", "refresh_token无法验证")
		} else {
			StdOutUnit.Info("", "无token可验证")
		}
		return "", StdOutUnit.GetErrorMessage(-403, "无法验证的令牌")
	}

	accessPre := strings.Split(token.AccessToken, ".")
	if len(accessPre) != 3 {
		StdOutUnit.Info("", "access_token格式错误")
		return "", StdOutUnit.GetErrorMessage(-403, "令牌无效")
	}
	headerPre, err := base64.StdEncoding.DecodeString(accessPre[1])
	if err != nil {
		StdOutUnit.Warn("", "token header解析错误", err)
		return "", StdOutUnit.GetErrorMessage(-403, "令牌无效")
	}
	header := strings.Split(strings.ReplaceAll(string(headerPre), "%", ""), "&")
	if len(header) != 2 {
		StdOutUnit.Info("", "token header格式错误")
		return "", StdOutUnit.GetErrorMessage(-403, "令牌无效")
	}
	username = header[0]
	password, errMessage := SessionManager.GetUserPassword(username, "")
	if errMessage.HasInfo {
		return "", errMessage
	}

	//IF DEBUG
	password, errMessage = RSAStaticUnit.DecodePublicEncode(password)
	if errMessage.HasInfo {
		return "", errMessage
	}
	if len(password) <= 8 {
		return "", StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	password = password[8:]
	//ENDIF

	if password == "" {
		return "", StdOutUnit.GetErrorMessage(-403, "令牌无效")
	}
	tokenCreateTime, err := strconv.ParseInt(header[1], 10, 64)
	if err != nil {
		StdOutUnit.Warn("", "token创建时间解析错误", err)
		return username, StdOutUnit.GetErrorMessage(-403, "令牌无效")
	}
	if tokenCreateTime+access < time.Now().Unix() {
		StdOutUnit.Info("", "access_token过期")
		if token.RefreshToken == "" {
			return username, StdOutUnit.GetErrorMessage(-403, "令牌失效")
		}
	}
	accessBodyPre, _ := json.Marshal(TokenBody{
		Password: password,
		Time:     tokenCreateTime,
		Type:     "access",
		Key:      tokenKey,
	})
	accessBody := getMD5(accessBodyPre)
	if accessBody != accessPre[0] {
		//StdOutUnit.Debug(username, "password: " + password, nil)
		StdOutUnit.Info("", "access_token body无效")
		return username, StdOutUnit.GetErrorMessage(-403, "令牌无效")
	}
	accessCheckPre := accessBody + "." + accessPre[1] + "." + tokenSecret
	if getMD5([]byte(accessCheckPre)) != accessPre[2] {
		StdOutUnit.Info("", "access_token签名无效")
		return username, StdOutUnit.GetErrorMessage(-403, "令牌无效")
	}

	if token.RefreshToken == "" {
		return username, StdOutUnit.GetEmptyErrorMessage()
	}
	if tokenCreateTime+refresh < time.Now().Unix() {
		StdOutUnit.Info("", "refresh_token过期")
		return username, StdOutUnit.GetErrorMessage(-403, "令牌失效")
	}
	refreshPre := strings.Split(token.RefreshToken, ".")
	if len(refreshPre) != 2 {
		StdOutUnit.Info("", "refresh_token格式错误")
		return username, StdOutUnit.GetErrorMessage(-403, "令牌无效")
	}
	refreshBodyPre, _ := json.Marshal(TokenBody{
		Password: password,
		Time:     tokenCreateTime,
		Type:     "refresh",
		Key:      tokenKey,
	})
	refreshBody := getMD5(refreshBodyPre)
	if refreshBody != refreshPre[0] {
		StdOutUnit.Info("", "refresh_token body无效")
		return username, StdOutUnit.GetErrorMessage(-403, "令牌无效")
	}
	refreshCheckPre := refreshBody + "." + accessPre[1] + "." + tokenSecret
	if getFullMD5([]byte(refreshCheckPre)) != refreshPre[1] {
		StdOutUnit.Info("", "refresh_token签名无效")
		return username, StdOutUnit.GetErrorMessage(-403, "令牌无效")
	} else {
		return username, StdOutUnit.GetEmptyErrorMessage()
	}
}

func getFullMD5(data []byte) string {
	h := md5.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func getMD5(data []byte) string {
	return getFullMD5(data)[8:24]
}
