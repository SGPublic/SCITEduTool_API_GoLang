package InfoModule

import (
	"SCITEduTool/consts"
	"SCITEduTool/manager/ChartManager"
	"SCITEduTool/manager/InfoManager"
	"SCITEduTool/module/SessionModule"
	"SCITEduTool/unit/StdOutUnit"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

func Get(username string) (InfoManager.UserInfo, StdOutUnit.MessagedError) {
	info, errMessage := InfoManager.Get(username)
	if errMessage.HasInfo {
		return InfoManager.UserInfo{}, errMessage
	}
	if !info.Exist {
		StdOutUnit.Info(username, "用户信息不存在")
		goto refresh
	}
	if !info.Expired {
		return info, StdOutUnit.GetEmptyErrorMessage()
	}
	StdOutUnit.Info(username, "用户基本信息过期")

refresh:
	session, identify, errMessage := SessionModule.Get(username, "")
	if errMessage.HasInfo {
		return InfoManager.UserInfo{}, errMessage
	}
	info, errMessage = Refresh(username, session, identify)
	if errMessage.HasInfo {
		return InfoManager.UserInfo{}, errMessage
	} else {
		return info, StdOutUnit.GetEmptyErrorMessage()
	}
}

func Refresh(username string, session string, identify int) (InfoManager.UserInfo, StdOutUnit.MessagedError) {
	switch identify {
	case 0:
		return studentInfo(username, session)
	case 1:
		return teacherInfo(username, session)
	default:
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
}

func studentInfo(username string, session string) (InfoManager.UserInfo, StdOutUnit.MessagedError) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	urlString := "http://218.6.163.93:8081/xsgrxx.aspx?xh=" + username
	req, _ := http.NewRequest("GET", urlString, nil)
	req.AddCookie(&http.Cookie{Name: "ASP.NET_SessionId", Value: session})
	req.Header.Add("Referer", urlString)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		StdOutUnit.Error(username, "网络请求失败", err)
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	resp.Body.Close()
	if err != nil {
		StdOutUnit.Error("", "HTML解析失败", err)
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	viewState := doc.Find("#__VIEWSTATE").AttrOr("value", "")
	if viewState == "" {
		StdOutUnit.Error(username, "未发现 __VIEWSTATE", nil)
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	gradePre := doc.Find("#lbl_dqszj").Text()
	if gradePre == "" {
		StdOutUnit.Error(username, "年级获取失败", nil)
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	grade, err := strconv.Atoi(gradePre)
	if err != nil {
		StdOutUnit.Error(username, "年级ID解析失败", err)
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	name := doc.Find("#xm").Text()
	if name == "" {
		StdOutUnit.Error(username, "姓名名称获取失败", nil)
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	lblXzb := doc.Find("#lbl_xzb").Text()
	if lblXzb == "" {
		StdOutUnit.Error(username, "班级名称获取失败", nil)
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	r, _ := regexp.Compile("(\\d+)\\.?(\\d+)班")
	classPre := r.FindString(lblXzb)
	if classPre == "" {
		StdOutUnit.Error(username, "班级ID获取失败", nil)
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	class, err := strconv.Atoi(strings.ReplaceAll(classPre, "班", ""))
	if err != nil {
		StdOutUnit.Error(username, "班级ID解析失败", nil)
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	lblXy := doc.Find("#lbl_xy").Text()
	if lblXy == "" {
		StdOutUnit.Error(username, "学院名称获取失败", nil)
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	lblZymc := doc.Find("#lbl_zymc").Text()
	StdOutUnit.Debug(username, lblZymc, nil)
	if lblZymc == "" {
		StdOutUnit.Error(username, "专业名称获取失败", nil)
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	urlString = "http://218.6.163.93:8081/tjkbcx.aspx?xh=" + username
	req, _ = http.NewRequest("GET", urlString, nil)
	req.AddCookie(&http.Cookie{Name: "ASP.NET_SessionId", Value: session})
	req.Header.Add("Referer", urlString)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err = client.Do(req)
	if err != nil {
		StdOutUnit.Error(username, "网络请求失败", err)
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	doc, err = goquery.NewDocumentFromReader(resp.Body)
	resp.Body.Close()
	if err != nil {
		StdOutUnit.Error("", "HTML解析失败", err)
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	viewState = doc.Find("#__VIEWSTATE").AttrOr("value", "")
	if viewState == "" {
		StdOutUnit.Error(username, "未发现 __VIEWSTATE", nil)
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	lblXyId := -1
	doc.Find("#xy").Find("option").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if s.Text() == lblXy {
			lblXyId, _ = strconv.Atoi(s.AttrOr("value", "-1"))
			return false
		}
		return true
	})
	if lblXyId == -1 {
		StdOutUnit.Error(username, "学院ID获取失败", nil)
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	form := url.Values{}
	form.Set("__EVENTTARGET", "xq")
	form.Set("__EVENTARGUMENT", "")
	form.Set("__LASTFOCUS", "")
	form.Set("__VIEWSTATE", viewState)
	form.Set("__VIEWSTATEGENERATOR", "3189F21D")
	form.Set("xn", consts.SchoolYear)
	form.Set("xq", "1")
	form.Set("nj", gradePre)
	form.Set("xy", strconv.Itoa(lblXyId))
	req, _ = http.NewRequest("POST", urlString, strings.NewReader(strings.TrimSpace(form.Encode())))
	req.AddCookie(&http.Cookie{Name: "ASP.NET_SessionId", Value: session})
	req.Header.Add("Referer", urlString)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err = client.Do(req)
	if err != nil {
		StdOutUnit.Error(username, "网络请求失败", err)
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	doc, err = goquery.NewDocumentFromReader(resp.Body)
	resp.Body.Close()
	if err != nil {
		StdOutUnit.Error("", "HTML解析失败", err)
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	lblZymcId := -1
	doc.Find("#zy").Find("option").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if s.Text() == lblZymc {
			lblZymcId, _ = strconv.Atoi(s.AttrOr("value", "-1"))
			return false
		}
		return true
	})
	if lblZymcId == -1 {
		StdOutUnit.Error(username, "专业ID获取失败", nil)
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	ChartManager.WriteFacultyName(lblXyId, lblXy)
	ChartManager.WriteSpecialtyName(lblXyId, lblZymcId, lblZymc)
	ChartManager.WriteClassName(lblXyId, lblZymcId, class, lblXzb)
	InfoManager.Update(username, name, lblXyId, lblZymcId, class, grade)
	return InfoManager.UserInfo{
		Name:      name,
		Faculty:   lblXyId,
		Specialty: lblZymcId,
		Class:     class,
		Grade:     grade,
	}, StdOutUnit.GetEmptyErrorMessage()
}

func teacherInfo(username string, session string) (InfoManager.UserInfo, StdOutUnit.MessagedError) {
	return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "TODO")
}
