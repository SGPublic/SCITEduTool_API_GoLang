package api

import (
	"net/http"

	"SCITEduTool/manager/SessionManager"
	"SCITEduTool/unit/TokenUnit"
)

func Token(w http.ResponseWriter, r *http.Request) {
	base, errMessage := SetupAPI(w, r, map[string]string{
		"access_token":  "",
		"refresh_token": "",
	})
	var token TokenUnit.Token
	var password string
	var refresh string
	var username string
	if errMessage.HasInfo {
		goto ouError
	}
	refresh = base.GetParameter("refresh_token")
	username, errMessage = TokenUnit.Check(TokenUnit.Token{
		AccessToken:  base.GetParameter("access_token"),
		RefreshToken: refresh,
	})
	if errMessage.HasInfo {
		goto ouError
	}

	password, errMessage = SessionManager.GetUserPassword(username, "")
	if errMessage.HasInfo {
		goto ouError
	}
	token, errMessage = TokenUnit.Build(username, password)
	if errMessage.HasInfo {
		goto ouError
	}
	base.OnObjectResult(struct {
		Code         int    `json:"code"`
		Message      string `json:"message"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}{
		Code:         200,
		Message:      "success.",
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	})
	return

ouError:
	errMessage.OutMessage(w)
}
