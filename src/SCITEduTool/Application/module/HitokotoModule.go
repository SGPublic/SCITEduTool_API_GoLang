package module

import (
	"SCITEduTool/Application/manager"
	"SCITEduTool/Application/stdio"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type hitokotoModule interface {
	Get() (manager.HitokotoItem, stdio.MessagedError)
	Refresh() (manager.HitokotoItem, stdio.MessagedError)
}

type hitokotoModuleImpl struct{}

var HitokotoModule hitokotoModule = hitokotoModuleImpl{}

func (hitokotoModuleImpl hitokotoModuleImpl) Get() (manager.HitokotoItem, stdio.MessagedError) {
	hitokoto, errMessage := manager.HitokotoManager.Get()
	if errMessage.HasInfo {
		return manager.HitokotoItem{}, errMessage
	}
	if !hitokoto.Exist {
		stdio.LogInfo("", "Hitokoto待更新")
		goto insert
	} else {
		return hitokoto, stdio.GetEmptyErrorMessage()
	}

insert:
	hitokoto, errMessage = HitokotoModule.Refresh()
	if !errMessage.HasInfo {
		return hitokoto, stdio.GetEmptyErrorMessage()
	} else {
		return manager.HitokotoItem{}, stdio.GetErrorMessage(-500, "请求处理失败")
	}
}

func (hitokotoModuleImpl hitokotoModuleImpl) Refresh() (manager.HitokotoItem, stdio.MessagedError) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, _ := http.NewRequest("GET", "https://v1.hitokoto.cn/?encode=json", nil)
	resp, err := client.Do(req)
	if err != nil {
		stdio.LogError("", "网络请求失败", err)
		return manager.HitokotoItem{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	body, _ := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	item := manager.HitokotoItem{}
	err = json.Unmarshal(body, &item)
	if err != nil {
		stdio.LogError("", "Hitokoto解析失败", err)
		return manager.HitokotoItem{}, stdio.GetErrorMessage(-500, "请求处理失败")
	}
	errMessage := manager.HitokotoManager.Insert(item)
	if errMessage.HasInfo {
		return manager.HitokotoItem{}, stdio.GetErrorMessage(-500, "请求处理失败")
	}
	return item, stdio.GetEmptyErrorMessage()
}
