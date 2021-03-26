package AchieveModule

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"SCITEduTool/base/LocalDebug"
	"SCITEduTool/manager/AchieveManager"
	"SCITEduTool/manager/InfoManager"
	"SCITEduTool/module/InfoModule"
	"SCITEduTool/module/SessionModule"
	"SCITEduTool/unit/StdOutUnit"
)

type ExtractTaskInfo struct {
	Username    string
	TaskID      int
	Year        string
	Semester    int
	TargetTasks []SingleTaskInfo
}

type TaskStatus struct {
	TaskID  int              `json:"task_id"`
	Success []SingleTaskInfo `json:"success"`
	Warn    []WarnTaskInfo   `json:"warn"`
	Failed  []FailedTaskInfo `json:"failed"`
}

type SingleTaskInfo struct {
	Username string `json:"uid"`
	Name     string `json:"name"`
}

type WarnTaskInfo struct {
	Username     string `json:"uid"`
	Name         string `json:"name"`
	NameInternal string `json:"name_internal"`
}

type FailedTaskInfo struct {
	Username  string `json:"uid"`
	Name      string `json:"name"`
	ErrorInfo string `json:"error_info"`
}

func ExtractPrepare(info ExtractTaskInfo) (TaskStatus, StdOutUnit.MessagedError) {
	status := TaskStatus{
		TaskID:  info.TaskID,
		Success: make([]SingleTaskInfo, 0),
		Warn:    make([]WarnTaskInfo, 0),
		Failed:  make([]FailedTaskInfo, 0),
	}
	if status.TaskID < 0 {
		status.TaskID = int(time.Now().Unix() / 300)
	}
	extractPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		StdOutUnit.Warn("", "运行目录获取失败", err)
		return status, StdOutUnit.GetErrorMessage(-500, "请求处理失败")
	}
	extractPath += "/achieve/extract/" + strconv.Itoa(status.TaskID) + "/" + info.Username + "/prepare/"
	if LocalDebug.IsDebug() {
		extractPath = strings.ReplaceAll(extractPath, "/", "\\")
	}
	_, err = os.Stat(extractPath)
	if err != nil {
		_ = os.MkdirAll(extractPath, 0644)
		//if err == os.ErrNotExist {
		//	_ = os.MkdirAll(extractPath, 0644)
		//} else {
		//	StdOutUnit.Warn("", "导出预备目录获取失败", err)
		//	return status, StdOutUnit.GetErrorMessage(-500, "请求处理失败")
		//}
	} else if info.TaskID < 0 {
		return status, StdOutUnit.GetErrorMessage(-500, "短时间内请勿多次创建导出任务")
	}
	for _, singleTask := range info.TargetTasks {
		if singleTask.Username == "" {
			status.Failed = append(status.Failed, FailedTaskInfo{
				Name:      singleTask.Name,
				Username:  singleTask.Username,
				ErrorInfo: "学号为空",
			})
			continue
		}
		data, errMessage := AchieveManager.Get(singleTask.Username, info.Year, info.Semester)
		if errMessage.HasInfo {
			status.Failed = append(status.Failed, FailedTaskInfo{
				Name:      singleTask.Name,
				Username:  singleTask.Username,
				ErrorInfo: data.ErrorInfo,
			})
			continue
		}
		err := ioutil.WriteFile(extractPath+singleTask.Username+".xlsx", data.Data, 0644)
		if err != nil {
			status.Failed = append(status.Failed, FailedTaskInfo{
				Name:      singleTask.Name,
				Username:  singleTask.Username,
				ErrorInfo: data.ErrorInfo,
			})
		} else if data.Name != singleTask.Name {
			status.Warn = append(status.Warn, WarnTaskInfo{
				Username:     singleTask.Username,
				Name:         singleTask.Name,
				NameInternal: data.Name,
			})
		} else {
			status.Success = append(status.Success, singleTask)
		}
	}
	return status, StdOutUnit.GetEmptyErrorMessage()
}

func ExtractFinal(info ExtractTaskInfo) StdOutUnit.MessagedError {
	extractPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		StdOutUnit.Warn("", "运行目录获取失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理失败")
	}
	extractPath += "/achieve/extract/" + strconv.Itoa(info.TaskID) + "/" + info.Username + "/"
	if LocalDebug.IsDebug() {
		extractPath = strings.ReplaceAll(extractPath, "/", "\\")
	}
	_, err = os.Stat(extractPath)
	if err != nil {
		StdOutUnit.Warn("", "导出预备目录获取失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理失败")
	}
	file, err := os.Open(extractPath)
	if err != nil {
		StdOutUnit.Warn("", "导出预备目录获取失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理失败")
	}

	zipFile, _ := os.Create(extractPath + "extract_" + strconv.Itoa(info.TaskID) + ".zip.prepare")
	extract := zip.NewWriter(zipFile)
	dirInside, err := file.ReadDir(-1)
	if err != nil {
		StdOutUnit.Warn("", "导出预备目录获取失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理失败")
	}
	for _, dirIndex := range dirInside {
		StdOutUnit.Debug(info.Username, dirIndex.Name(), nil)
		//if dirIndex.IsDir() {
		//	continue
		//}
		var singleFile *os.File
		if LocalDebug.IsDebug() {
			singleFile, err = os.Open(extractPath + "prepare\\" + dirIndex.Name())
		} else {
			singleFile, err = os.Open(extractPath + "prepare/" + dirIndex.Name())
		}
		if err != nil {
			StdOutUnit.Warn("", "成绩单读取失败", err)
			continue
		}
		info, err := singleFile.Stat()
		if err != nil {
			StdOutUnit.Warn("", "成绩单读取失败", err)
			continue
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			StdOutUnit.Warn("", "成绩单读取失败", err)
			continue
		}
		header.Method = zip.Store
		writer, err := extract.CreateHeader(header)
		if err != nil {
			StdOutUnit.Warn("", "成绩单读取失败", err)
			continue
		}
		_, err = io.Copy(writer, singleFile)
		if err != nil {
			StdOutUnit.Warn("", "成绩单读取失败", err)
		}
		singleFile.Close()
	}
	extract.Close()
	zipFile.Close()
	file.Close()
	err = os.Rename(extractPath+"extract_"+strconv.Itoa(info.TaskID)+".zip.prepare", extractPath+"extract_"+strconv.Itoa(info.TaskID)+".zip")
	if err != nil {
		StdOutUnit.Warn("", "成绩单读取失败", err)
	}
	return StdOutUnit.GetEmptyErrorMessage()
}

func Get(username string, year string, semester int) (AchieveManager.AchieveObject,
	StdOutUnit.MessagedError) {
	session, _, errMessage := SessionModule.Get(username, "")
	if errMessage.HasInfo {
		return AchieveManager.AchieveObject{}, errMessage
	}
	info, errMessage := InfoModule.Get(username)
	if errMessage.HasInfo {
		return AchieveManager.AchieveObject{}, errMessage
	}
	tableContent, errMessage := Refresh(username, year, semester, session, info)
	if errMessage.HasInfo {
		return AchieveManager.AchieveObject{}, errMessage
	} else {
		return tableContent, StdOutUnit.GetEmptyErrorMessage()
	}
}

func Refresh(username string, year string, semester int, session string,
	info InfoManager.UserInfo) (AchieveManager.AchieveObject,
	StdOutUnit.MessagedError) {
	switch info.Identify {
	case 0:
		return studentAchieve(username, year, semester, session, info)
	case 1:
		return teacherAchieve()
	default:
		return AchieveManager.AchieveObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
}

func studentAchieve(username string, year string, semester int, session string,
	info InfoManager.UserInfo) (AchieveManager.AchieveObject,
	StdOutUnit.MessagedError) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	Button1 := "按学期查询"
	if year == "all" {
		Button1 = "在校学习成绩查询"
	} else if semester == 0 {
		Button1 = "按学年查询"
	}
	urlString := "http://218.6.163.93:8081/xscj.aspx?xh=" + username
	req, _ := http.NewRequest("GET", urlString, nil)
	req.AddCookie(&http.Cookie{Name: "ASP.NET_SessionId", Value: session})
	req.Header.Add("Referer", urlString)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		StdOutUnit.Error(username, "网络请求失败", err)
		return AchieveManager.AchieveObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	r, _ := regexp.Compile("__VIEWSTATE\" value=\"(.*?)\"")
	viewState := r.FindString(string(body))
	if viewState == "" {
		StdOutUnit.Error(username, "未发现 __VIEWSTATE", nil)
		return AchieveManager.AchieveObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	viewState = viewState[20 : len(viewState)-1]

	form := url.Values{}
	form.Set("__VIEWSTATE", viewState)
	form.Set("__VIEWSTATEGENERATOR", "17EB693E")
	form.Set("ddlXN", year)
	form.Set("ddlXQ", strconv.Itoa(semester))
	form.Set("txtQSCJ", "0")
	form.Set("txtZZCJ", "100")
	form.Set("Button1", Button1)
	req, _ = http.NewRequest("POST", urlString, strings.NewReader(strings.TrimSpace(form.Encode())))
	req.AddCookie(&http.Cookie{Name: "ASP.NET_SessionId", Value: session})
	req.Header.Add("Referer", urlString)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err = client.Do(req)
	if err != nil {
		StdOutUnit.Error(username, "网络请求失败", err)
		return AchieveManager.AchieveObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	body, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	r, _ = regexp.Compile("__VIEWSTATE\" value=\"(.*?)\"")
	viewState = r.FindString(string(body))
	if viewState == "" {
		StdOutUnit.Error(username, "未发现 __VIEWSTATE", nil)
		return AchieveManager.AchieveObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	bodyString := string(body)
	bodyString = strings.ReplaceAll(bodyString, "\n", "")
	bodyString = strings.ReplaceAll(bodyString, " class=\"alt\"", "")

	achieveObject := AchieveManager.AchieveObject{
		Current: []AchieveManager.CurrentAchieveItem{},
		Failed:  []AchieveManager.FailedAchieveItem{},
	}

	var currentMatch string
	var currentMatches []string
	var failedMatch string
	var failedMatches []string
	r, _ = regexp.Compile("id=\"DataGrid1\"(.*?)</table>")
	if !r.MatchString(bodyString) {
		StdOutUnit.Info(username, "用户目标学期无成绩单")
		goto next
	}
	bodyString = strings.ReplaceAll(bodyString, "&nbsp;", "")
	currentMatch = r.FindString(bodyString)
	currentMatch = currentMatch[14 : len(currentMatch)-8]

	r, _ = regexp.Compile("<tr>(.*?)</tr>")
	currentMatches = r.FindAllString(currentMatch, -1)
	r, _ = regexp.Compile("<td>(.*?)</td>")
	for _, currentItem := range currentMatches {
		if !r.MatchString(currentItem) {
			continue
		}
		explodeGradeIndex := r.FindAllString(currentItem, -1)
		currentAchieveItem := AchieveManager.CurrentAchieveItem{}
		currentAchieveItem.Name = explodeGradeIndex[1][4 : len(explodeGradeIndex[1])-5]
		currentAchieveItem.PaperScore = explodeGradeIndex[3][4 : len(explodeGradeIndex[3])-5]
		currentAchieveItem.Mark = explodeGradeIndex[4][4 : len(explodeGradeIndex[4])-5]
		currentAchieveItem.Retake = explodeGradeIndex[6][4 : len(explodeGradeIndex[6])-5]
		currentAchieveItem.Rebuild = explodeGradeIndex[7][4 : len(explodeGradeIndex[7])-5]
		currentAchieveItem.Credit = explodeGradeIndex[8][4 : len(explodeGradeIndex[8])-5]
		achieveObject.Current = append(achieveObject.Current, currentAchieveItem)
	}

next:
	r, _ = regexp.Compile("id=\"Datagrid3\"(.*?)</table>")
	if !r.MatchString(bodyString) {
		StdOutUnit.Info(username, "用户无挂科")
		goto result
	}
	failedMatch = r.FindString(bodyString)
	failedMatch = failedMatch[14 : len(failedMatch)-8]

	r, _ = regexp.Compile("<tr>(.*?)</tr>")
	failedMatches = r.FindAllString(failedMatch, -1)
	r, _ = regexp.Compile("<td>(.*?)</td>")
	for _, failedItem := range failedMatches {
		if !r.MatchString(failedItem) {
			continue
		}
		explodeGradeIndex := r.FindAllString(failedItem, -1)
		failedAchieveItem := AchieveManager.FailedAchieveItem{}
		failedAchieveItem.Name = explodeGradeIndex[1][4 : len(explodeGradeIndex[1])-5]
		failedAchieveItem.Mark = explodeGradeIndex[3][4 : len(explodeGradeIndex[3])-5]
		achieveObject.Failed = append(achieveObject.Failed, failedAchieveItem)
	}

result:
	AchieveManager.Update(username, info, year, semester, achieveObject)
	return achieveObject, StdOutUnit.GetEmptyErrorMessage()
}

func teacherAchieve() (AchieveManager.AchieveObject, StdOutUnit.MessagedError) {
	return AchieveManager.AchieveObject{}, StdOutUnit.GetErrorMessage(-500, "什么？老师还有成绩单？(°Д°≡°Д°)")
}
