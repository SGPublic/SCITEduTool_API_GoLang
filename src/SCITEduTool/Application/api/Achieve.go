package api

import (
	"SCITEduTool/Application/consts"
	"SCITEduTool/Application/manager"
	"SCITEduTool/Application/module"
	"SCITEduTool/Application/stdio"
	"net/http"
	"strconv"
)

func Achieve(w http.ResponseWriter, r *http.Request) {
	base, errMessage := SetupAPI(w, r, map[string]string{
		"access_token": "",
		"semester":     strconv.Itoa(consts.Semester),
		"year":         consts.SchoolYear,
	})
	if errMessage.HasInfo {
		errMessage.OutMessage(w)
		return
	}
	accessToken := base.GetParameter("access_token")
	username, errMessage := manager.TokenUnit.Check(manager.Token{
		AccessToken: accessToken,
	})
	if errMessage.HasInfo {
		errMessage.OutMessage(w)
		return
	}

	semester, err := strconv.Atoi(base.GetParameter("semester"))
	if err != nil {
		stdio.LogInfo(username, "学期参数解析失败")
		base.OnStandardMessage(-500, "请求处理出错")
		return
	}
	year := base.GetParameter("year")

	achieve, errMessage := module.AchieveModule.Get(username, year, semester)
	if errMessage.HasInfo {
		errMessage.OutMessage(w)
		return
	}
	base.OnObjectResult(struct {
		Code    int                   `json:"code"`
		Message string                `json:"message"`
		Achieve manager.AchieveObject `json:"achieve"`
	}{
		Code:    200,
		Message: "success.",
		Achieve: achieve,
	})
}
