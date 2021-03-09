package SessionHelper

import (
	"SCITEduTool/manager/SessionManager"
	"SCITEduTool/unit/RSAStaticUnit"
	"SCITEduTool/unit/StdOutUnit"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func Get(username string, password string) (string, int, StdOutUnit.MessagedError) {
	sessionExists, err := SessionManager.Get(username)
	if err.HasInfo {
		return "", 0, err
	}
	if !sessionExists.Exist {
		StdOutUnit.Info.String(username, "用户 ASP.NET_SessionId 不存在")
		return "", 0, StdOutUnit.GetEmptyErrorMessage()
	}
	if !sessionExists.Effective {
		StdOutUnit.Info.String(username, "用户修改密码，登陆状态失效")
		return "", 0, StdOutUnit.GetErrorMessage(-401, "登录状态失效")
	}
	if !sessionExists.Expired {
		return sessionExists.Session, sessionExists.Identify, StdOutUnit.GetEmptyErrorMessage()
	}
	StdOutUnit.Info.String(username, "用户 ASP.NET_SessionId 过期")
	password, err = SessionManager.GetUserPassword(username, password)
	if err.HasInfo || password == "" {
		return "", 0, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	session, identify, err := Refresh(username, password)
	if !err.HasInfo {
		return session, identify, StdOutUnit.GetEmptyErrorMessage()
	}
	if err.Code == -401 && password == "" {
		return "", 0, StdOutUnit.GetErrorMessage(-401, "登陆状态失效，请重新登录")
	} else {
		return "", 0, err
	}
}

func Refresh(username string, password string) (string, int, StdOutUnit.MessagedError) {
	location, identify, errMessage := GetVerifyLocation(username, password)
	if errMessage.HasInfo {
		return "", 0, errMessage
	}
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, _ := http.NewRequest("GET", location, nil)
	resp, err := client.Do(req)
	if err != nil {
		StdOutUnit.Error.String(username, err.Error())
		return "", 0, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	cookies := resp.Header.Values("Set-Cookie")
	r, _ := regexp.Compile("ASP.NET_SessionId=(.*?);")
	session := ""
	for _, cookie := range cookies {
		session = r.FindString(cookie)
		if session != "" {
			break
		}
	}
	if session == "" {
		StdOutUnit.Error.String(username, "ASP.NET_SessionId 获取失败")
		return "", 0, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	if len(session) <= 18 {
		StdOutUnit.Error.String(username, "ASP.NET_SessionId 处理失败")
		return "", 0, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	session = session[18 : len(session)-1]
	StdOutUnit.Verbose.String(username, "用户获取 ASP.NET_SessionId 成功")
	SessionManager.Update(username, password, session, identify)
	return session, identify, StdOutUnit.GetEmptyErrorMessage()
}

func GetVerifyLocation(username string, password string) (string, int, StdOutUnit.MessagedError) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, _ := http.NewRequest("GET", "http://218.6.163.95:18080/zfca/login", nil)
	resp, err := client.Do(req)
	if err != nil {
		StdOutUnit.Error.String(username, err.Error())
		return "", 0, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	cookies := resp.Header.Values("Set-Cookie")
	r, _ := regexp.Compile("JSESSIONID=(.*?);")
	Jsessionid1 := ""
	for _, cookie := range cookies {
		Jsessionid1 = r.FindString(cookie)
		if Jsessionid1 != "" {
			break
		}
	}
	if Jsessionid1 == "" {
		StdOutUnit.Error.String(username, "JSESSIONID1 获取失败")
		return "", 0, StdOutUnit.GetErrorMessage(-401, "账号或密码错误")
	}
	if len(Jsessionid1) <= 11 {
		StdOutUnit.Error.String(username, "JSESSIONID1 处理失败")
		return "", 0, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	Jsessionid1 = Jsessionid1[11 : len(Jsessionid1)-1]

	body, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	r, _ = regexp.Compile("lt\" value=\"(.*?)\"")
	lt := r.FindString(string(body))
	if len(lt) <= 11 {
		StdOutUnit.Error.String(username, "lt 获取失败")
		return "", 0, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	lt = lt[11 : len(lt)-1]

	passwordDecode := ""
	decode, errDecode := RSAStaticUnit.DecodePublicEncode(password)
	if errDecode.HasInfo {
		return "", 0, errDecode
	}
	if len(decode) <= 8 {
		return "", 0, StdOutUnit.GetErrorMessage(-401, "密码校验失败")
	}
	passwordDecode = decode[8:]

	form := url.Values{}
	form.Set("useValidateCode", "0")
	form.Set("isremenberme", "0")
	form.Set("ip", "")
	form.Set("username", username)
	form.Set("password", passwordDecode)
	form.Set("losetime", "30")
	form.Set("lt", lt)
	form.Set("_eventId", "submit")
	form.Set("submit1", "+")
	req, err = http.NewRequest("POST", "http://218.6.163.95:18080/zfca/login;jsessionid="+
		Jsessionid1, strings.NewReader(strings.TrimSpace(form.Encode())))
	if err != nil {
		StdOutUnit.Error.String(username, err.Error())
		return "", 0, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err = client.Do(req)
	if err != nil {
		StdOutUnit.Error.String(username, err.Error())
		return "", 0, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	cookies = resp.Header.Values("Set-Cookie")
	r, _ = regexp.Compile("CASTGC=(.*?);")
	castgc := ""
	for _, cookie := range cookies {
		castgc = r.FindString(cookie)
		if castgc != "" {
			break
		}
	}
	if castgc == "" {
		StdOutUnit.Info.String(username, "登录账号或密码错误")
		return "", 0, StdOutUnit.GetErrorMessage(-401, "账号或密码错误")
	}
	if len(castgc) <= 7 {
		StdOutUnit.Error.String(username, "CASTGC 获取失败")
		return "", 0, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	castgc = castgc[7 : len(castgc)-1]
	location := resp.Header.Values("Location")
	if len(location) != 1 {
		StdOutUnit.Error.String(username, "第一次跳转失败")
		return "", 0, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	req, _ = http.NewRequest("GET", location[0], nil)
	resp, err = client.Do(req)
	if err != nil {
		StdOutUnit.Error.String(username, err.Error())
		return "", 0, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	cookies = resp.Header.Values("Set-Cookie")
	r, _ = regexp.Compile("JSESSIONID=(.*?);")
	Jsessionid2 := ""
	for _, cookie := range cookies {
		Jsessionid2 = r.FindString(cookie)
		if Jsessionid2 != "" {
			break
		}
	}
	if Jsessionid2 == "" {
		return "", 0, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	if len(Jsessionid2) <= 11 {
		return "", 0, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	Jsessionid2 = Jsessionid2[11 : len(Jsessionid2)-1]

	location = resp.Header.Values("Location")
	if len(location) != 1 {
		StdOutUnit.Error.String(username, "第二次跳转失败")
		return "", 0, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	req, err = http.NewRequest("GET", location[0], nil)
	if err != nil {
		StdOutUnit.Error.String(username, err.Error())
		return "", 0, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	req.AddCookie(&http.Cookie{Name: "JSESSIONID", Value: Jsessionid1})
	req.AddCookie(&http.Cookie{Name: "CASTGC", Value: castgc})
	req.AddCookie(&http.Cookie{Name: "JSESSIONID", Value: Jsessionid2})
	resp, err = client.Do(req)
	if err != nil {
		StdOutUnit.Error.String(username, err.Error())
		return "", 0, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	location = resp.Header.Values("Location")
	if len(location) != 1 {
		StdOutUnit.Error.String(username, "第三次跳转失败")
		return "", 0, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	req, err = http.NewRequest("GET", location[0], nil)
	if err != nil {
		StdOutUnit.Error.String(username, err.Error())
		return "", 0, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	req.AddCookie(&http.Cookie{Name: "JSESSIONID", Value: Jsessionid2})
	resp, err = client.Do(req)
	if err != nil {
		StdOutUnit.Error.String(username, err.Error())
		return "", 0, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	body, err = ioutil.ReadAll(resp.Body)
	identity := -1
	student, _ := regexp.MatchString("student", string(body))
	teacher, _ := regexp.MatchString("teacher", string(body))
	resp.Body.Close()
	switch true {
	case student:
		identity = 0
		break
	case teacher:
		identity = 1
		break
	default:
		StdOutUnit.Error.String(username, "identity 获取失败")
		return "", 0, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	identities := []string{"student", "teacher"}
	location = []string{
		"http://218.6.163.95:18080/zfca/login?yhlx=" + identities[identity] +
			"&login=0122579031373493708&url=xs_main.aspx",
	}
	req, err = http.NewRequest("GET", location[0], nil)
	if err != nil {
		StdOutUnit.Error.String(username, ""+err.Error())
		return "", 0, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	req.AddCookie(&http.Cookie{Name: "JSESSIONID", Value: Jsessionid1})
	req.AddCookie(&http.Cookie{Name: "CASTGC", Value: castgc})
	req.AddCookie(&http.Cookie{Name: "JSESSIONID", Value: Jsessionid2})
	resp, err = client.Do(req)
	if err != nil {
		StdOutUnit.Error.String(username, ""+err.Error())
		return "", 0, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	location = resp.Header.Values("Location")
	if len(location) != 1 {
		StdOutUnit.Error.String(username, "跳转链接获取失败")
		return "", 0, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	StdOutUnit.Verbose.String(username, "用户获取跳转链接成功")
	return location[0], identity, StdOutUnit.GetEmptyErrorMessage()
}
