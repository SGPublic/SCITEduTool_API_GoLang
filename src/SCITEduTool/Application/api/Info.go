package api

import (
	"SCITEduTool/Application/manager"
	"SCITEduTool/Application/module"
	base2 "SCITEduTool/Application/stdio"
	"net/http"
	"strconv"
)

type InfoOut struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Info    InfoOutContent `json:"info"`
}

type InfoOutContent struct {
	Name      string `json:"name"`
	Identify  int    `json:"identify"`
	Level     int    `json:"level"`
	Faculty   string `json:"faculty"`
	Specialty string `json:"specialty"`
	Class     string `json:"class"`
	Grade     string `json:"grade"`
}

func Info(w http.ResponseWriter, r *http.Request) {
	base, err := SetupAPI(w, r, map[string]string{
		"access_token": "",
	})
	if err.HasInfo {
		err.OutMessage(w)
		return
	}
	accessToken := base.GetParameter("access_token")
	username, err := manager.TokenUnit.Check(manager.Token{
		AccessToken: accessToken,
	})
	if err.HasInfo {
		err.OutMessage(w)
		return
	}
	info, err := module.InfoModule.Get(username)
	if err.HasInfo {
		err.OutMessage(w)
		return
	}
	var ChartManager = manager.ChartManager
	faculty, err := ChartManager.GetFacultyName(info.Faculty)
	if err.HasInfo {
		err.OutMessage(w)
		return
	}
	specialty, err := ChartManager.GetSpecialtyName(info.Faculty, info.Specialty)
	if err.HasInfo {
		err.OutMessage(w)
		return
	}
	class, err := ChartManager.GetClassName(info.Faculty, info.Specialty, info.Class)
	if err.HasInfo {
		err.OutMessage(w)
		return
	}

	base2.LogVerbose(username, "用户获取基本信息成功")
	base.OnObjectResult(InfoOut{
		Code:    200,
		Message: "success.",
		Info: InfoOutContent{
			Name:      info.Name,
			Identify:  info.Identify,
			Level:     info.Level,
			Faculty:   faculty,
			Specialty: specialty,
			Class:     class,
			Grade:     strconv.Itoa(info.Grade),
		},
	})
}
