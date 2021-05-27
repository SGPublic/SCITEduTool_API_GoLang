package module

import (
	"SCITEduTool/Application/consts"
	"SCITEduTool/Application/manager"
	"SCITEduTool/Application/stdio"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type infoModule interface {
	Get(username string) (manager.UserInfo, stdio.MessagedError)
	Refresh(username string, session string, identify int) (manager.UserInfo, stdio.MessagedError)
}

type infoModuleImpl struct{}

var InfoModule infoModule = infoModuleImpl{}

func (infoModuleImpl infoModuleImpl) Get(username string) (manager.UserInfo, stdio.MessagedError) {
	info, errMessage := manager.InfoManager.Get(username)
	if errMessage.HasInfo {
		return manager.UserInfo{}, errMessage
	}
	if !info.Exist {
		stdio.LogInfo(username, "用户信息不存在")
		goto refresh
	}
	if !info.Expired {
		return info, stdio.GetEmptyErrorMessage()
	}
	stdio.LogInfo(username, "用户基本信息过期")

refresh:
	session, identify, errMessage := SessionModule.Get(username, "")
	if errMessage.HasInfo {
		return manager.UserInfo{}, errMessage
	}
	info, errMessage = InfoModule.Refresh(username, session, identify)
	if errMessage.HasInfo {
		return manager.UserInfo{}, errMessage
	} else {
		return info, stdio.GetEmptyErrorMessage()
	}
}

func (infoModuleImpl infoModuleImpl) Refresh(username string, session string, identify int) (manager.UserInfo, stdio.MessagedError) {
	switch identify {
	case 0:
		return studentInfo(username, session)
	case 1:
		return teacherInfo(username, session)
	default:
		return manager.UserInfo{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
}

func studentInfo(username string, session string) (manager.UserInfo, stdio.MessagedError) {
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
		stdio.LogError(username, "网络请求失败", err)
		return manager.UserInfo{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	resp.Body.Close()
	if err != nil {
		stdio.LogError("", "HTML解析失败", err)
		return manager.UserInfo{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	viewState := doc.Find("#__VIEWSTATE").AttrOr("value", "")
	if viewState == "" {
		stdio.LogError(username, "未发现 __VIEWSTATE", nil)
		return manager.UserInfo{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}

	gradePre := doc.Find("#lbl_dqszj").Text()
	if gradePre == "" {
		stdio.LogError(username, "年级获取失败", nil)
		return manager.UserInfo{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	grade, err := strconv.Atoi(gradePre)
	if err != nil {
		stdio.LogError(username, "年级ID解析失败", err)
		return manager.UserInfo{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}

	name := doc.Find("#xm").Text()
	if name == "" {
		stdio.LogError(username, "姓名名称获取失败", nil)
		return manager.UserInfo{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}

	lblXzb := doc.Find("#lbl_xzb").Text()
	if lblXzb == "" {
		stdio.LogError(username, "班级名称获取失败", nil)
		return manager.UserInfo{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	r, _ := regexp.Compile("(\\d+)\\.?(\\d+)班")
	classPre := r.FindString(lblXzb)
	if classPre == "" {
		stdio.LogError(username, "班级ID获取失败", nil)
		return manager.UserInfo{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	class, err := strconv.Atoi(strings.ReplaceAll(classPre, "班", ""))
	if err != nil {
		stdio.LogError(username, "班级ID解析失败", nil)
		return manager.UserInfo{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}

	lblXy := doc.Find("#lbl_xy").Text()
	if lblXy == "" {
		stdio.LogError(username, "学院名称获取失败", nil)
		return manager.UserInfo{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}

	lblZymc := doc.Find("#lbl_zymc").Text()
	stdio.LogDebug(username, lblZymc, nil)
	if lblZymc == "" {
		stdio.LogError(username, "专业名称获取失败", nil)
		return manager.UserInfo{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}

	urlString = "http://218.6.163.93:8081/tjkbcx.aspx?xh=" + username
	req, _ = http.NewRequest("GET", urlString, nil)
	req.AddCookie(&http.Cookie{Name: "ASP.NET_SessionId", Value: session})
	req.Header.Add("Referer", urlString)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err = client.Do(req)
	if err != nil {
		stdio.LogError(username, "网络请求失败", err)
		return manager.UserInfo{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	doc, err = goquery.NewDocumentFromReader(resp.Body)
	resp.Body.Close()
	if err != nil {
		stdio.LogError("", "HTML解析失败", err)
		return manager.UserInfo{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	viewState = doc.Find("#__VIEWSTATE").AttrOr("value", "")
	if viewState == "" {
		stdio.LogError(username, "未发现 __VIEWSTATE", nil)
		return manager.UserInfo{}, stdio.GetErrorMessage(-500, "请求处理出错")
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
		stdio.LogError(username, "学院ID获取失败", nil)
		return manager.UserInfo{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}

	yearStart, _ := strconv.Atoi(strings.Split(consts.SchoolYear, "-")[0])
	lblZymcId := -1
	for i := 0; i > -6; i-- {
		year := strconv.Itoa(yearStart+i) + "-" + strconv.Itoa(yearStart+1+i)
		form := url.Values{}
		form.Set("__EVENTTARGET", "xq")
		form.Set("__EVENTARGUMENT", "")
		form.Set("__LASTFOCUS", "")
		form.Set("__VIEWSTATE", viewState)
		form.Set("__VIEWSTATEGENERATOR", "3189F21D")
		form.Set("xn", year)
		form.Set("xq", "1")
		form.Set("nj", gradePre)
		form.Set("xy", strconv.Itoa(lblXyId))
		req, _ = http.NewRequest("POST", urlString, strings.NewReader(strings.TrimSpace(form.Encode())))
		req.AddCookie(&http.Cookie{Name: "ASP.NET_SessionId", Value: session})
		req.Header.Add("Referer", urlString)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err = client.Do(req)
		if err != nil {
			stdio.LogError(username, "网络请求失败", err)
			return manager.UserInfo{}, stdio.GetErrorMessage(-500, "请求处理出错")
		}

		doc, err = goquery.NewDocumentFromReader(resp.Body)
		resp.Body.Close()
		if err != nil {
			stdio.LogError("", "HTML解析失败", err)
			return manager.UserInfo{}, stdio.GetErrorMessage(-500, "请求处理出错")
		}

		doc.Find("#zy").Find("option").EachWithBreak(func(i int, s *goquery.Selection) bool {
			if s.Text() == lblZymc {
				lblZymcId, _ = strconv.Atoi(s.AttrOr("value", "-1"))
				return false
			}
			return true
		})
		if lblZymcId != -1 {
			break
		}
	}
	if lblZymcId == -1 {
		stdio.LogError(username, "专业ID获取失败", nil)
		return manager.UserInfo{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}

	manager.ChartManager.WriteFacultyName(lblXyId, lblXy)
	manager.ChartManager.WriteSpecialtyName(lblXyId, lblZymcId, lblZymc)
	manager.ChartManager.WriteClassName(lblXyId, lblZymcId, class, lblXzb)
	manager.InfoManager.Update(username, name, lblXyId, lblZymcId, class, grade)
	return manager.UserInfo{
		Name:      name,
		Faculty:   lblXyId,
		Specialty: lblZymcId,
		Class:     class,
		Grade:     grade,
	}, stdio.GetEmptyErrorMessage()
}

func teacherInfo(username string, session string) (manager.UserInfo, stdio.MessagedError) {
	return manager.UserInfo{}, stdio.GetErrorMessage(-500, "TODO")
}
