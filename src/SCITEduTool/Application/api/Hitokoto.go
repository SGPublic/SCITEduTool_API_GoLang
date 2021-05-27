package api

import (
	"SCITEduTool/Application/module"
	"net/http"
)

func Hitokoto(w http.ResponseWriter, r *http.Request) {
	base, errMessage := SetupAPI(w, r, nil)
	if errMessage.HasInfo {
		errMessage.OutMessage(w)
		return
	}
	item, errMessage := module.HitokotoModule.Get()
	if errMessage.HasInfo {
		errMessage.OutMessage(w)
		return
	}
	base.OnObjectResult(struct {
		Code     int    `json:"code"`
		Message  string `json:"message"`
		Hitokoto string `json:"hitokoto"`
		From     string `json:"from"`
	}{
		Code:     200,
		Message:  "success.",
		Hitokoto: item.Content,
		From:     item.From,
	})
}
