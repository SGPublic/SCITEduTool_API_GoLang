package api

import (
	"SCITEduTool/Application/consts"
	"SCITEduTool/Application/manager"
	"SCITEduTool/Application/module"
	base2 "SCITEduTool/Application/stdio"
	"net/http"
	"strconv"
)

func Table(w http.ResponseWriter, r *http.Request) {
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
		base2.LogInfo(username, "学期参数解析失败")
		base.OnStandardMessage(-500, "请求处理出错")
		return
	}
	year := base.GetParameter("year")

	table, errMessage := module.TableModule.Get(username, year, semester)
	if errMessage.HasInfo {
		errMessage.OutMessage(w)
		return
	}
	base.OnObjectResult(struct {
		Code    int                      `json:"code"`
		Message string                   `json:"message"`
		Table   [7][5]manager.LessonItem `json:"table"`
	}{
		Code:    200,
		Message: "success.",
		Table:   table.Object,
	})
}
