package module

import (
	"SCITEduTool/Application/stdio"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

type examModule interface {
	Get(username string) (ExamObject, stdio.MessagedError)
	Refresh(username string, session string, identify int) (ExamObject, stdio.MessagedError)
}

type examModuleImpl struct{}

var ExamModule examModule = examModuleImpl{}

type ExamItem struct {
	Name     string `json:"name"`
	Time     string `json:"time"`
	Location string `json:"location"`
	SetNum   string `json:"set_num"`
}

type ExamObject struct {
	Object []ExamItem `json:"exam"`
}

func (examModuleImpl examModuleImpl) Get(username string) (ExamObject, stdio.MessagedError) {
	session, identify, errMessage := SessionModule.Get(username, "")
	if errMessage.HasInfo {
		return ExamObject{}, errMessage
	}
	examContent, errMessage := ExamModule.Refresh(username, session, identify)
	if errMessage.HasInfo {
		return ExamObject{}, errMessage
	} else {
		return examContent, stdio.GetEmptyErrorMessage()
	}
}

func (examModuleImpl examModuleImpl) Refresh(username string, session string, identify int) (ExamObject,
	stdio.MessagedError) {
	switch identify {
	case 0:
		return studentExam(username, session)
	case 1:
		return teacherExam()
	default:
		return ExamObject{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
}

func studentExam(username string, session string) (ExamObject,
	stdio.MessagedError) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	urlString := "http://218.6.163.93:8081/xskscx.aspx?xh=" + username
	req, _ := http.NewRequest("GET", urlString, nil)
	req.AddCookie(&http.Cookie{Name: "ASP.NET_SessionId", Value: session})
	req.Header.Add("Referer", urlString)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		stdio.LogError(username, "网络请求失败", err)
		return ExamObject{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	r, _ := regexp.Compile("__VIEWSTATE\" value=\"(.*?)\"")
	viewState := r.FindString(string(body))
	if viewState == "" {
		stdio.LogError(username, "未发现 __VIEWSTATE", nil)
		return ExamObject{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}

	bodyString := string(body)
	bodyString = strings.ReplaceAll(bodyString, "\n", "")
	bodyString = strings.ReplaceAll(bodyString, " class=\"alt\"", "")

	examObject := ExamObject{
		Object: []ExamItem{},
	}

	var examMatch string
	var examMatches []string
	r, _ = regexp.Compile("id=\"DataGrid1\"(.*?)</table>")
	if !r.MatchString(bodyString) {
		stdio.LogInfo(username, "用户目标学期无成绩单")
		goto result
	}
	bodyString = strings.ReplaceAll(bodyString, "&nbsp;", "")
	examMatch = r.FindString(bodyString)
	examMatch = examMatch[14 : len(examMatch)-8]

	r, _ = regexp.Compile("<tr>(.*?)</tr>")
	examMatches = r.FindAllString(examMatch, -1)
	r, _ = regexp.Compile("<td>(.*?)</td>")
	for _, currentItem := range examMatches {
		if !r.MatchString(currentItem) {
			continue
		}
		explodeExamIndex := r.FindAllString(currentItem, -1)
		currentExamItem := ExamItem{}
		currentExamItem.Name = explodeExamIndex[1][4 : len(explodeExamIndex[1])-5]
		currentExamItem.Time = explodeExamIndex[3][4 : len(explodeExamIndex[3])-5]
		currentExamItem.Location = explodeExamIndex[4][4 : len(explodeExamIndex[4])-5]
		currentExamItem.SetNum = explodeExamIndex[6][4 : len(explodeExamIndex[6])-5]
		examObject.Object = append(examObject.Object, currentExamItem)
	}

result:
	stdio.LogVerbose(username, "用户获取考试安排信息成功")
	return examObject, stdio.GetEmptyErrorMessage()
}

func teacherExam() (ExamObject, stdio.MessagedError) {
	return ExamObject{}, stdio.GetErrorMessage(-500, "TODO")
}
