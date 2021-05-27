package api

import (
	"SCITEduTool/Application/manager"
	"SCITEduTool/Application/module"
	"net/http"
)

func Springboard(w http.ResponseWriter, r *http.Request) {
	base, err := SetupAPI(w, r, map[string]string{
		"access_token": "",
	})
	if err.HasInfo {
		err.OutMessage(w)
		return
	}
	username, err := manager.TokenUnit.Check(manager.Token{
		AccessToken: base.GetParameter("access_token"),
	})
	if err.HasInfo {
		err.OutMessage(w)
		return
	}
	password, err := manager.SessionManager.GetUserPassword(username, "")
	if err.HasInfo {
		err.OutMessage(w)
		return
	}
	location, _, err := module.SessionModule.GetVerifyLocation(username, password)
	if err.HasInfo {
		err.OutMessage(w)
		return
	}
	base.OnObjectResult(struct {
		Code     int    `json:"code"`
		Message  string `json:"message"`
		Location string `json:"location"`
	}{
		Code:     200,
		Message:  "success.",
		Location: location,
	})
}
