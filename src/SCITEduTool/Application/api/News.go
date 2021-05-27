package api

import (
	"SCITEduTool/Application/manager"
	"SCITEduTool/Application/module"
	"net/http"
	"strconv"
)

func News(w http.ResponseWriter, r *http.Request) {
	base, errMessage := SetupAPI(w, r, map[string]string{
		"action": "",
		"tid":    "-1",
		"page":   "-1",
	})
	if errMessage.HasInfo {
		errMessage.OutMessage(w)
		return
	}
	action := base.GetParameter("action")
	switch action {
	case "type":
		Type(w, base)
		break
	case "list":
		List(w, base)
		break
	//case "get":
	//	break
	default:
		Headline(w, base)
		break
	}
}

func Type(w http.ResponseWriter, api BaseAPI) {
	charts, errMessage := module.NewsModule.GetTypeChart()
	if errMessage.HasInfo {
		errMessage.OutMessage(w)
		return
	}
	var chartsOut []manager.NewsTypeChartItem
	for _, item := range charts {
		if item.Out != 1 {
			continue
		}
		chartsOut = append(chartsOut, item)
	}
	api.OnObjectResult(struct {
		Code    int                         `json:"code"`
		Message string                      `json:"message"`
		Charts  []manager.NewsTypeChartItem `json:"charts"`
	}{
		Code:    200,
		Message: "success.",
		Charts:  chartsOut,
	})
}

func List(w http.ResponseWriter, api BaseAPI) {
	tidPre := api.GetParameter("tid")
	tid, err := strconv.Atoi(tidPre)
	if err != nil || tid < 0 {
		api.OnStandardMessage(-500, "无效的参数")
		return
	}
	pagePre := api.GetParameter("page")
	page, err := strconv.Atoi(pagePre)
	if err != nil || page < 0 {
		api.OnStandardMessage(-500, "无效的参数")
		return
	}
	news, hasNext, errMessage := module.NewsModule.ListNewsByType(tid, page)
	if errMessage.HasInfo {
		errMessage.OutMessage(w)
		return
	}
	api.OnObjectResult(struct {
		Code    int                `json:"code"`
		Message string             `json:"message"`
		HasNext bool               `json:"has_next"`
		News    []manager.NewsItem `json:"news"`
	}{
		Code:    200,
		Message: "success.",
		HasNext: hasNext,
		News:    news,
	})
}

func Headline(w http.ResponseWriter, api BaseAPI) {
	headlines, errMessage := module.NewsModule.GetHeadlines()
	if errMessage.HasInfo {
		errMessage.OutMessage(w)
		return
	}
	api.OnObjectResult(struct {
		Code      int                `json:"code"`
		Message   string             `json:"message"`
		Headlines []manager.NewsItem `json:"headlines"`
	}{
		Code:      200,
		Message:   "success.",
		Headlines: headlines,
	})
}
