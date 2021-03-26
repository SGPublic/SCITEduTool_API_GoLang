package api

import (
	"SCITEduTool/consts"
	"net/http"
	"strconv"

	"SCITEduTool/manager/TableManager"
	"SCITEduTool/module/TableModule"
	"SCITEduTool/unit/StdOutUnit"
	"SCITEduTool/unit/TokenUnit"
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
	username, errMessage := TokenUnit.Check(TokenUnit.Token{
		AccessToken: accessToken,
	})
	if errMessage.HasInfo {
		errMessage.OutMessage(w)
		return
	}

	semester, err := strconv.Atoi(base.GetParameter("semester"))
	if err != nil {
		StdOutUnit.Info(username, "学期参数解析失败")
		StdOutUnit.GetErrorMessage(-500, "请求处理出错").OutMessage(w)
		return
	}
	year := base.GetParameter("year")

	table, errMessage := TableModule.Get(username, year, semester)
	if errMessage.HasInfo {
		errMessage.OutMessage(w)
		return
	}
	base.OnObjectResult(struct {
		Code    int                           `json:"code"`
		Message string                        `json:"message"`
		Table   [6][5]TableManager.LessonItem `json:"table"`
	}{
		Code:    200,
		Message: "success.",
		Table:   table.Object,
	})
}
