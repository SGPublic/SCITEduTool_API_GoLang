package module

import (
	"SCITEduTool/Application/manager"
	"SCITEduTool/Application/stdio"
	"SCITEduTool/Application/unit"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type sessionModule interface {
	Get(username string, password string) (string, int, stdio.MessagedError)
	Refresh(username string, password string) (string, int, stdio.MessagedError)
	GetVerifyLocation(username string, password string) (string, int, stdio.MessagedError)
}

type sessionModuleImpl struct{}

var SessionModule sessionModule = sessionModuleImpl{}

func (sessionModuleImpl sessionModuleImpl) Get(username string, password string) (string, int, stdio.MessagedError) {
	sessionExists, err := manager.SessionManager.Get(username)
	if err.HasInfo {
		return "", 0, err
	}
	if !sessionExists.Exist {
		goto refresh
	}
	if !sessionExists.Effective {
		stdio.LogInfo(username, "用户修改密码，登陆状态失效")
		return "", 0, stdio.GetErrorMessage(-401, "登录状态失效")
	}
	if !sessionExists.Expired {
		return sessionExists.Session, sessionExists.Identify, stdio.GetEmptyErrorMessage()
	}
	stdio.LogInfo(username, "用户 ASP.NET_SessionId 过期")

refresh:
	session, identify, err := SessionModule.Refresh(username, password)
	if !err.HasInfo {
		return session, identify, stdio.GetEmptyErrorMessage()
	}
	if err.Code == -401 && password == "" {
		return "", 0, stdio.GetErrorMessage(-401, "登陆状态失效，请重新登录")
	} else {
		return "", 0, err
	}
}

func (sessionModuleImpl sessionModuleImpl) Refresh(username string, password string) (string, int, stdio.MessagedError) {
	var errMessage stdio.MessagedError
	if password == "" {
		password, errMessage = manager.SessionManager.GetUserPassword(username, "")
		if errMessage.HasInfo {
			return "", 0, errMessage
		}
	}
	location, identify, errMessage := SessionModule.GetVerifyLocation(username, password)
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
		stdio.LogError(username, "网络请求失败", err)
		return "", 0, stdio.GetErrorMessage(-500, "请求处理出错")
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
		stdio.LogError(username, "ASP.NET_SessionId 获取失败", nil)
		return "", 0, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	if len(session) <= 18 {
		stdio.LogError(username, "ASP.NET_SessionId 处理失败", nil)
		return "", 0, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	session = session[18 : len(session)-1]
	stdio.LogVerbose(username, "用户获取 ASP.NET_SessionId 成功")
	manager.SessionManager.Update(username, password, session, identify)
	return session, identify, stdio.GetEmptyErrorMessage()
}

func (sessionModuleImpl sessionModuleImpl) GetVerifyLocation(username string, password string) (string, int, stdio.MessagedError) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, _ := http.NewRequest("GET", "http://218.6.163.95:18080/zfca/login", nil)
	resp, err := client.Do(req)
	if err != nil {
		stdio.LogError(username, "网络请求失败", err)
		return "", 0, stdio.GetErrorMessage(-500, "请求处理出错")
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
		stdio.LogError(username, "JSESSIONID1 获取失败", nil)
		return "", 0, stdio.GetErrorMessage(-401, "账号或密码错误")
	}
	if len(Jsessionid1) <= 11 {
		stdio.LogError(username, "JSESSIONID1 处理失败", nil)
		return "", 0, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	Jsessionid1 = Jsessionid1[11 : len(Jsessionid1)-1]

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	resp.Body.Close()
	if err != nil {
		stdio.LogError("", "HTML解析失败", err)
		return "", 0, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	ltDoc := doc.Find(".btn").Find("span").Find("input")
	if ltDoc.AttrOr("name", "nil") != "lt" {
		stdio.LogError(username, "lt 获取失败", nil)
		return "", 0, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	lt := ltDoc.AttrOr("value", "nil")
	if lt == "nil" {
		stdio.LogError(username, "lt 解析失败", nil)
		return "", 0, stdio.GetErrorMessage(-500, "请求处理出错")
	}

	passwordDecode := password
	//IF !DEBUG
	decode, errDecode := unit.RSAStaticUnit.DecodePublicEncode(passwordDecode)
	if errDecode.HasInfo {
		return "", 0, errDecode
	}
	if len(decode) <= 8 {
		return "", 0, stdio.GetErrorMessage(-401, "密码校验失败")
	}
	passwordDecode = decode[8:]
	//ENDIF

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
	req, _ = http.NewRequest("POST", "http://218.6.163.95:18080/zfca/login;jsessionid="+
		Jsessionid1, strings.NewReader(strings.TrimSpace(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err = client.Do(req)
	if err != nil {
		stdio.LogError(username, "网络请求失败", err)
		return "", 0, stdio.GetErrorMessage(-500, "请求处理出错")
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
		stdio.LogInfo(username, "登录账号或密码错误")
		return "", 0, stdio.GetErrorMessage(-401, "账号或密码错误")
	}
	if len(castgc) <= 7 {
		stdio.LogError(username, "CASTGC 获取失败", nil)
		return "", 0, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	castgc = castgc[7 : len(castgc)-1]
	location := resp.Header.Values("Location")
	if len(location) != 1 {
		stdio.LogError(username, "第一次跳转失败", nil)
		return "", 0, stdio.GetErrorMessage(-500, "请求处理出错")
	}

	req, _ = http.NewRequest("GET", location[0], nil)
	resp, err = client.Do(req)
	if err != nil {
		stdio.LogError(username, "请求失败创建", err)
		return "", 0, stdio.GetErrorMessage(-500, "请求处理出错")
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
		return "", 0, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	if len(Jsessionid2) <= 11 {
		return "", 0, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	Jsessionid2 = Jsessionid2[11 : len(Jsessionid2)-1]

	location = resp.Header.Values("Location")
	if len(location) != 1 {
		stdio.LogError(username, "第二次跳转失败", nil)
		return "", 0, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	req, _ = http.NewRequest("GET", location[0], nil)
	req.AddCookie(&http.Cookie{Name: "JSESSIONID", Value: Jsessionid1})
	req.AddCookie(&http.Cookie{Name: "CASTGC", Value: castgc})
	req.AddCookie(&http.Cookie{Name: "JSESSIONID", Value: Jsessionid2})
	resp, err = client.Do(req)
	if err != nil {
		stdio.LogError(username, "网络请求失败", err)
		return "", 0, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	location = resp.Header.Values("Location")
	if len(location) != 1 {
		stdio.LogError(username, "第三次跳转失败", nil)
		return "", 0, stdio.GetErrorMessage(-500, "请求处理出错")
	}

	req, _ = http.NewRequest("GET", location[0], nil)
	req.AddCookie(&http.Cookie{Name: "JSESSIONID", Value: Jsessionid2})
	resp, err = client.Do(req)
	if err != nil {
		stdio.LogError(username, "网络请求失败", err)
		return "", 0, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	body, err := ioutil.ReadAll(resp.Body)
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
		stdio.LogError(username, "identity 获取失败", nil)
		return "", 0, stdio.GetErrorMessage(-500, "请求处理出错")
	}

	identities := []string{"student", "teacher"}
	location = []string{
		"http://218.6.163.95:18080/zfca/login?yhlx=" + identities[identity] +
			"&login=0122579031373493708&url=xs_main.aspx",
	}
	req, _ = http.NewRequest("GET", location[0], nil)
	req.AddCookie(&http.Cookie{Name: "JSESSIONID", Value: Jsessionid1})
	req.AddCookie(&http.Cookie{Name: "CASTGC", Value: castgc})
	req.AddCookie(&http.Cookie{Name: "JSESSIONID", Value: Jsessionid2})
	resp, err = client.Do(req)
	if err != nil {
		stdio.LogError(username, "网络请求失败", err)
		return "", 0, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	location = resp.Header.Values("Location")
	if len(location) != 1 {
		stdio.LogError(username, "跳转链接获取失败", nil)
		return "", 0, stdio.GetErrorMessage(-500, "请求处理出错")
	}

	stdio.LogVerbose(username, "用户获取跳转链接成功")
	return location[0], identity, stdio.GetEmptyErrorMessage()
}
