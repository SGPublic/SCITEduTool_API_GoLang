package api

import (
	"SCITEduTool/helper/InfoHelper"
	"SCITEduTool/manager/ChartManager"
	"SCITEduTool/unit/StdOutUnit"
	"SCITEduTool/unit/TokenUnit"
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
	username, err := TokenUnit.Check(TokenUnit.Token{
		AccessToken: accessToken,
	})
	if err.HasInfo {
		err.OutMessage(w)
		return
	}
	info, err := InfoHelper.Get(username)
	if err.HasInfo {
		err.OutMessage(w)
		return
	}
	faculty, err := ChartManager.GetFacultyName(info.Faculty)
	if err.HasInfo {
		err.OutMessage(w)
		return
	}
	specialty, err := ChartManager.GetFacultyName(info.Specialty)
	if err.HasInfo {
		err.OutMessage(w)
		return
	}
	class, err := ChartManager.GetFacultyName(info.Class)
	if err.HasInfo {
		err.OutMessage(w)
		return
	}

	StdOutUnit.OnObjectResult(w, InfoOut{
		Code:    200,
		Message: "success.",
		Info: InfoOutContent{
			Name:      info.Name,
			Faculty:   faculty,
			Specialty: specialty,
			Class:     class,
			Grade:     strconv.Itoa(info.Grade),
		},
	})
}
