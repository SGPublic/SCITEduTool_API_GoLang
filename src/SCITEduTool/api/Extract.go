package api

import (
	"SCITEduTool/consts"
	"SCITEduTool/module/AchieveModule"
	"SCITEduTool/module/InfoModule"
	"SCITEduTool/unit/StdOutUnit"
	"SCITEduTool/unit/TokenUnit"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func Extract(w http.ResponseWriter, r *http.Request) {
	base, errMessage := SetupAPI(w, r, map[string]string{
		"access_token": "",
		"task_id":      "-1",
		"semester":     strconv.Itoa(consts.Semester),
		"year":         consts.SchoolYear,
		"tasks":        "",
	})
	if errMessage.HasInfo {
		errMessage.OutMessage(w)
		return
	}
	accessToken := base.GetParameter("access_token")
	username, errMessage := TokenUnit.Check(TokenUnit.Token{
		AccessToken: accessToken,
	})
	if errMessage.HasInfo {
		errMessage.OutMessage(w)
		return
	}
	taskId, err := strconv.Atoi(base.GetParameter("task_id"))
	if err != nil {
		base.OnStandardMessage(-500, "无效的参数")
		return
	}
	semester, err := strconv.Atoi(base.GetParameter("semester"))
	if err != nil {
		base.OnStandardMessage(-500, "无效的参数")
		return
	}
	tasksPre, err := base64.StdEncoding.DecodeString(base.GetParameter("tasks"))
	if err != nil {
		StdOutUnit.Debug(username, "tasks参数解析失败", err)
		base.OnStandardMessage(-500, "无效的参数")
		return
	}
	var tasks []AchieveModule.SingleTaskInfo
	err = json.Unmarshal(tasksPre, &tasks)
	if err != nil {
		StdOutUnit.Debug(username, "tasks参数解析失败", err)
		base.OnStandardMessage(-500, "无效的参数")
		return
	}
	if len(tasks) > 100 {
		base.OnStandardMessage(-500, "请勿一次性提交过量的任务")
		return
	}
	info, errMessage := InfoModule.Get(username)
	if errMessage.HasInfo {
		errMessage.OutMessage(w)
		return
	}
	if info.Identify != 1 && info.Level < 80 {
		base.OnStandardMessage(-403, "权限不足")
	}
	status, errMessage := AchieveModule.ExtractPrepare(AchieveModule.ExtractTaskInfo{
		Username:    username,
		TaskID:      taskId,
		Year:        base.GetParameter("year"),
		Semester:    semester,
		TargetTasks: tasks,
	})
	if errMessage.HasInfo {
		errMessage.OutMessage(w)
		return
	}
	base.OnObjectResult(struct {
		Code    int                      `json:"code"`
		Message string                   `json:"message"`
		Status  AchieveModule.TaskStatus `json:"status"`
	}{
		Code:    200,
		Message: "success.",
		Status:  status,
	})
}

func ExtractDone(w http.ResponseWriter, r *http.Request) {
	base, errMessage := SetupAPI(w, r, map[string]string{
		"access_token": "",
		"task_id":      "-1",
	})
	if errMessage.HasInfo {
		errMessage.OutMessage(w)
		return
	}
	accessToken := base.GetParameter("access_token")
	username, errMessage := TokenUnit.Check(TokenUnit.Token{
		AccessToken: accessToken,
	})
	if errMessage.HasInfo {
		errMessage.OutMessage(w)
		return
	}
	taskId, err := strconv.Atoi(base.GetParameter("task_id"))
	if err != nil {
		base.OnStandardMessage(-500, "无效的参数")
		return
	}
	info, errMessage := InfoModule.Get(username)
	if errMessage.HasInfo {
		errMessage.OutMessage(w)
		return
	}
	if info.Identify != 1 && info.Level < 80 {
		base.OnStandardMessage(-403, "权限不足")
	}

	extractPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		StdOutUnit.Warn("", "运行目录获取失败", err)
		base.OnStandardMessage(-500, "请求处理失败")
	}
	extractPath += "/achieve/extract/" + strconv.Itoa(taskId) + "/" + username + "/"
	//IF DEBUG
	//	extractPath = strings.ReplaceAll(extractPath, "/", "\\")
	//ENDIF
	_, err = os.Stat(extractPath + "extract_" + strconv.Itoa(taskId) + ".zip.prepare")
	if err == nil {
		base.OnStandardMessage(201, "处理中，请稍后再试")
		return
	}
	_, err = os.Stat(extractPath + "extract_" + strconv.Itoa(taskId) + ".zip")
	if err != nil {
		go AchieveModule.ExtractFinal(AchieveModule.ExtractTaskInfo{
			Username: username,
			TaskID:   taskId,
		})
		base.OnStandardMessage(201, "已提交任务，请稍后再次访问")
		return
	}
	link := AchieveModule.ExtractLink(AchieveModule.ExtractTaskInfo{
		Username: username,
		TaskID:   taskId,
	}, accessToken)
	base.OnObjectResult(struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Link    string `json:"link"`
	}{
		Code:    200,
		Message: "success.",
		Link:    link,
	})
}

func ExtractDownload(w http.ResponseWriter, r *http.Request) {
	base, errMessage := SetupAPI(w, r, map[string]string{
		"access_token": "",
		"task_id":      "-1",
	})

	if errMessage.HasInfo {
		errMessage.OutMessage(w)
		return
	}
	accessToken := base.GetParameter("access_token")
	username, errMessage := TokenUnit.Check(TokenUnit.Token{
		AccessToken: accessToken,
	})
	if errMessage.HasInfo {
		errMessage.OutMessage(w)
		return
	}
	taskId, err := strconv.Atoi(base.GetParameter("task_id"))
	if err != nil {
		base.OnStandardMessage(-500, "无效的参数")
		return
	}
	info, errMessage := InfoModule.Get(username)
	if errMessage.HasInfo {
		errMessage.OutMessage(w)
		return
	}
	if info.Identify != 1 && info.Level < 80 {
		base.OnStandardMessage(-403, "权限不足")
	}

	extractPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		StdOutUnit.Warn("", "运行目录获取失败", err)
		base.OnStandardMessage(-500, "请求处理失败")
	}
	extractPath += "/achieve/extract/" + strconv.Itoa(taskId) + "/" + username + "/"
	//IF DEBUG
	//	extractPath = strings.ReplaceAll(extractPath, "/", "\\")
	//ENDIF
	_, err = os.Stat(extractPath + "extract_" + strconv.Itoa(taskId) + ".zip")
	if err != nil {
		base.OnStandardMessage(-500, "请求处理失败")
		return
	}
	//StdOutUnit.Warn(username, extractPath + "extract_" + strconv.Itoa(taskId) + ".zip", nil)
	data, err := ioutil.ReadFile(extractPath + "extract_" + strconv.Itoa(taskId) + ".zip")
	if err != nil {
		base.OnStandardMessage(-500, "请求处理失败")
		return
	}
	w.Header().Set("Content-Type", "application/x-zip-compressed")
	w.Header().Set("Content-Disposition", "attachment;filename=extract_"+strconv.Itoa(taskId)+".zip")
	_, _ = w.Write(data)
}
