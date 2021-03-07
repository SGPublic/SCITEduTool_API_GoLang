package TokenUnit

import (
	"SCITEduTool/base/LocalDebug"
	"SCITEduTool/manager/InfoManager"
	"SCITEduTool/unit/RSAStaticUnit"
	"SCITEduTool/unit/StdOutUnit"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

const (
	access int64 = 2592000 // 一个月 30 * 24 * 3600
	refresh int64 = 124416000 // 四年 30 * 24 * 3600 * 12 * 4
)

var tokenKey string
var tokenSecret string

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

func Setup(key string, secret string) {
	tokenKey = key
	tokenSecret = secret
}

func Build(username string, password string) (Token, StdOutUnit.MessagedError) {
	token := Token {
		AccessToken: "",
		RefreshToken: "",
	}
	timeNow := time.Now().Unix()
	headerPre := username + "&" + strconv.FormatInt(timeNow, 10)
	for len(headerPre) % 3 != 0 {
		headerPre += "%"
	}
	header := base64.StdEncoding.EncodeToString([]byte(headerPre))
	token.AccessToken += header + "."

	passwordDecoded := ""
	if LocalDebug.IsDebug() {
		passwordDecoded = password
	} else {
		passwordPre, err := RSAStaticUnit.DecodePublicEncode(password)
		if err.HasInfo {
			return Token{}, err
		}
		passwordDecoded = passwordPre
	}
	accessBodyPre, _ := json.Marshal(TokenBody{
		Password: passwordDecoded,
		Time:     timeNow,
		Type:     "access",
		Key:      tokenKey,
	})
	refreshBodyPre, _ := json.Marshal(TokenBody{
		Password: passwordDecoded,
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

func Check(token Token) (TokenCheckResult, StdOutUnit.MessagedError) {
	username := ""
	accessEffective := false
	refreshEffective := false
	if token.AccessToken != "" {
		accessPre := strings.Split(token.AccessToken, ".")
		if len(accessPre) != 3 {
			StdOutUnit.Debug.String("", "access_token格式错误")
			return TokenCheckResult{}, StdOutUnit.GetErrorMessage(-403, "令牌无效")
		}
		headerPre, err := base64.StdEncoding.DecodeString(accessPre[1])
		if err != nil {
			StdOutUnit.Debug.String("", "token header解析错误")
			return TokenCheckResult{}, StdOutUnit.GetErrorMessage(-403, "令牌无效")
		}
		header := strings.Split(strings.ReplaceAll(string(headerPre), "%", ""), "&")
		if len(header) != 2 {
			StdOutUnit.Debug.String("", "token header格式错误")
			return TokenCheckResult{}, StdOutUnit.GetErrorMessage(-403, "令牌无效")
		}
		username = header[0]
		password, message := InfoManager.GetUserPassword(username)
		if message.HasInfo {
			StdOutUnit.Debug.String("", "access_token用户不存在")
			return TokenCheckResult{}, StdOutUnit.GetErrorMessage(-403, "令牌无效")
		}
		tokenCreateTime, err := strconv.ParseInt(header[1], 10, 64)
		if err != nil {
			StdOutUnit.Debug.String("", "token创建时间解析错误")
			return TokenCheckResult{}, StdOutUnit.GetErrorMessage(-403, "令牌无效")
		}
		if tokenCreateTime + access < time.Now().Unix() {
			StdOutUnit.Debug.String("", "access_token过期")
			return TokenCheckResult{}, StdOutUnit.GetErrorMessage(-403, "令牌失效")
		}
		accessBodyPre, _ := json.Marshal(TokenBody{
			Password: password,
			Time:     tokenCreateTime,
			Type:     "access",
			Key:      tokenKey,
		})
		accessBody := getMD5(accessBodyPre)
		if accessBody != accessPre[0] {
			StdOutUnit.Debug.String("", "access_token body无效")
			return TokenCheckResult{}, StdOutUnit.GetErrorMessage(-403, "令牌无效")
		}
		accessCheckPre := accessBody + "." + accessPre[1] + "." + tokenSecret
		if getMD5([]byte(accessCheckPre)) != accessPre[2] {
			StdOutUnit.Debug.String("", "access_token签名无效")
			return TokenCheckResult{}, StdOutUnit.GetErrorMessage(-403, "令牌无效")
		} else {
			accessEffective = true
		}

		if token.RefreshToken != "" {
			if tokenCreateTime + refresh < time.Now().Unix() {
				StdOutUnit.Debug.String("", "refresh_token过期")
				return TokenCheckResult{}, StdOutUnit.GetErrorMessage(-403, "令牌失效")
			}
			refreshPre := strings.Split(token.RefreshToken, ".")
			if len(refreshPre) != 2 {
				StdOutUnit.Debug.String("", "refresh_token格式错误")
				return TokenCheckResult{}, StdOutUnit.GetErrorMessage(-403, "令牌无效")
			}
			refreshBodyPre, _ := json.Marshal(TokenBody{
				Password: password,
				Time:     tokenCreateTime,
				Type:     "refresh",
				Key:      tokenKey,
			})
			refreshBody := getMD5(refreshBodyPre)
			if refreshBody != refreshPre[0] {
				StdOutUnit.Debug.String("", "refresh_token body无效")
				return TokenCheckResult{}, StdOutUnit.GetErrorMessage(-403, "令牌无效")
			}
			refreshCheckPre := refreshBody + "." + accessPre[1] + "." + tokenSecret
			if getFullMD5([]byte(refreshCheckPre)) != refreshPre[1] {
				StdOutUnit.Debug.String("", "refresh_token签名无效")
				return TokenCheckResult{}, StdOutUnit.GetErrorMessage(-403, "令牌无效")
			} else {
				refreshEffective = true
			}
		}

	} else if token.RefreshToken != "" {
		return TokenCheckResult{}, StdOutUnit.GetErrorMessage(-403, "无法验证的令牌")
	}

	return TokenCheckResult{
		Username: username,
		AccessEffective: accessEffective,
		RefreshEffective: refreshEffective,
	}, StdOutUnit.GetEmptyErrorMessage()
}

func getFullMD5(data []byte) string {
	h := md5.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func getMD5(data []byte) string {
	return getFullMD5(data)[8:24]
}