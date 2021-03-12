package HitokotoModule

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"SCITEduTool/manager/HitokotoManager"
	"SCITEduTool/unit/StdOutUnit"
)

func Get() (HitokotoManager.HitokotoItem, StdOutUnit.MessagedError) {
	hitokoto, errMessage := HitokotoManager.Get()
	if errMessage.HasInfo {
		return HitokotoManager.HitokotoItem{}, errMessage
	}
	if !hitokoto.Exist {
		StdOutUnit.Info("", "Hitokoto待更新")
		goto insert
	} else {
		return hitokoto, StdOutUnit.GetEmptyErrorMessage()
	}

insert:
	hitokoto, errMessage = Refresh()
	if !errMessage.HasInfo {
		return hitokoto, StdOutUnit.GetEmptyErrorMessage()
	} else {
		return HitokotoManager.HitokotoItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理失败")
	}
}

func Refresh() (HitokotoManager.HitokotoItem, StdOutUnit.MessagedError) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, _ := http.NewRequest("GET", "https://v1.hitokoto.cn/?encode=json", nil)
	resp, err := client.Do(req)
	if err != nil {
		StdOutUnit.Error("", "网络请求失败", err)
		return HitokotoManager.HitokotoItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	body, _ := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	item := HitokotoManager.HitokotoItem{}
	err = json.Unmarshal(body, &item)
	if err != nil {
		StdOutUnit.Error("", "Hitokoto解析失败", err)
		return HitokotoManager.HitokotoItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理失败")
	}
	errMessage := HitokotoManager.Insert(item)
	if errMessage.HasInfo {
		return HitokotoManager.HitokotoItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理失败")
	}
	return item, StdOutUnit.GetEmptyErrorMessage()
}
