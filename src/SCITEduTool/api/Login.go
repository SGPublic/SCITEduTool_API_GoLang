package api

import (
	"SCITEduTool/helper/SessionHelper"
	"SCITEduTool/unit/StdOutUnit"
	"SCITEduTool/unit/TokenUnit"
	"net/http"
)

type LoginOut struct {
	Code         int    `json:"code"`
	Message      string `json:"message"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	base, err := SetupAPI(w, r, map[string]string{
		"username": "",
		"password": "",
	})
	if err.HasInfo {
		err.OutMessage(w)
		return
	}

	username := base.GetParameter("username")
	password := base.GetParameter("password")
	_, _, err = SessionHelper.Get(username, password)
	if err.HasInfo {
		err.OutMessage(w)
		return
	}
	token, err := TokenUnit.Build(username, password)
	if err.HasInfo {
		err.OutMessage(w)
		return
	}
	StdOutUnit.Verbose.String(username, "用户登录成功")
	StdOutUnit.OnObjectResult(w, LoginOut{
		Code:         200,
		Message:      "success.",
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	})
}
