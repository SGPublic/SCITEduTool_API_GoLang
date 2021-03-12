package api

import (
	"SCITEduTool/module/SessionModule"
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
	_, _, err = SessionModule.Get(username, password)
	if err.HasInfo {
		if err.Code == 401 {
			base.OnObjectResult(LoginOut{
				Code:    200,
				Message: "账号或密码错误",
			})
		} else {
			err.OutMessage(w)
		}
		return
	}
	token, err := TokenUnit.Build(username, password)
	if err.HasInfo {
		goto outError
	}
	StdOutUnit.Verbose(username, "用户登录成功")
	base.OnObjectResult(LoginOut{
		Code:         200,
		Message:      "success.",
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	})
	return

outError:
	err.OutMessage(w)
}
