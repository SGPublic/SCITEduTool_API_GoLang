package manager

import (
	"SCITEduTool/Application/stdio"
	"SCITEduTool/Application/unit"
	"database/sql"
	"encoding/json"
	"time"
)

type newsManager interface {
	GetNewsById(tid int, nid int) (NewsItem, stdio.MessagedError)
	UpdateNews(item NewsItem) stdio.MessagedError
	GetHeadlines() (Headlines, stdio.MessagedError)
	UpdateHeadlines(headlines []NewsItem) stdio.MessagedError
	CheckNewsExist(tid int, nid int) (bool, stdio.MessagedError)
	GetTypeChart() ([]NewsTypeChartItem, stdio.MessagedError)
	UpdateTypeChart(chart []NewsTypeChartItem) stdio.MessagedError
	CheckChartExist(nTypeId int) (bool, stdio.MessagedError)
}

type newsManagerImpl struct{}

var NewsManager newsManager = newsManagerImpl{}

type NewsItem struct {
	Tid        int      `json:"tid"`
	Nid        int      `json:"nid"`
	Images     []string `json:"images"`
	Title      string   `json:"title"`
	Summary    string   `json:"summary"`
	CreateTime string   `json:"create_time"`
}

type Headlines struct {
	Exist   bool
	Expired bool
	News    []NewsItem
}

type NewsTypeChartItem struct {
	TypeId   int    `json:"id"`
	TypeName string `json:"name"`
	Out      int    `json:"-"`
}

func (newsManagerImpl newsManagerImpl) GetNewsById(tid int, nid int) (NewsItem, stdio.MessagedError) {
	tx, err := unit.Maria.Begin()
	if err != nil {
		stdio.LogWarn("", "数据库开始事务失败", err)
		return NewsItem{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	state, err := tx.Prepare("select `n_images`,`n_title`,`n_summary`,`n_create_time` from `news` where `n_id`=? and `n_type_id`=?")
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn("", "数据库准备SQL指令失败", err)
		return NewsItem{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	rows := state.QueryRow(nid, tid)
	item := NewsItem{}
	images := ""
	err = rows.Scan(&images, &item.Title, &item.Summary, &item.CreateTime)
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn("", "数据库SQL指令执行失败", err)
		return NewsItem{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	tx.Commit()
	err = json.Unmarshal([]byte(images), &item)
	if err != nil {
		stdio.LogWarn("", "新闻图片数据解析失败", err)
		return NewsItem{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	item.Nid = nid
	item.Tid = tid
	return item, stdio.GetEmptyErrorMessage()
}

func (newsManagerImpl newsManagerImpl) UpdateNews(item NewsItem) stdio.MessagedError {
	exist, errMessage := NewsManager.CheckNewsExist(item.Tid, item.Nid)
	if errMessage.HasInfo {
		return errMessage
	}
	tx, err := unit.Maria.Begin()
	if err != nil {
		stdio.LogWarn("", "数据库开始事务失败", err)
		return stdio.GetErrorMessage(-500, "请求处理出错")
	}
	var state *sql.Stmt
	if exist {
		return stdio.GetEmptyErrorMessage()
	}
	state, err = tx.Prepare("insert into `news` (`n_id`, `n_type_id`, `n_title`, `n_summary`, `n_images`, `n_create_time`) values (?, ?, ?, ?, ?, ?)")
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn("", "数据库准备SQL指令失败", err)
		return stdio.GetErrorMessage(-500, "请求处理出错")
	}
	img, _ := json.Marshal(struct {
		Images []string `json:"images"`
	}{
		Images: item.Images,
	})
	_, err = state.Exec(item.Nid, item.Tid, item.Title, item.Summary, string(img), item.CreateTime)
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn("", "数据库SQL指令执行失败", err)
		return stdio.GetErrorMessage(-500, "请求处理出错")
	} else {
		tx.Commit()
		stdio.LogVerbose("", "向数据库插入新新闻成功")
		return stdio.GetEmptyErrorMessage()
	}
}

func (newsManagerImpl newsManagerImpl) GetHeadlines() (Headlines, stdio.MessagedError) {
	tx, err := unit.Maria.Begin()
	if err != nil {
		stdio.LogWarn("", "数据库开始事务失败", err)
		return Headlines{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	state, err := tx.Prepare("select `h_id`,`h_type_id`,`h_image`,`h_expired` from `news_headline` order by `h_id` desc")
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn("", "数据库准备SQL指令失败", err)
		return Headlines{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	rows, err := state.Query()
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Commit()
			stdio.LogInfo("", "头条新闻数据不存在")
			return Headlines{}, stdio.GetEmptyErrorMessage()
		} else {
			_ = tx.Rollback()
			stdio.LogWarn("", "数据库SQL指令执行失败", err)
			return Headlines{}, stdio.GetErrorMessage(-500, "请求处理出错")
		}
	}
	expired := false
	var items []NewsItem
	for rows.Next() {
		item := NewsItem{}
		var itemExpired int64
		image := ""
		err = rows.Scan(&item.Nid, &item.Tid, &image, &itemExpired)
		if err != nil {
			if err == sql.ErrNoRows {
				tx.Commit()
				stdio.LogInfo("", "头条新闻详情不存在")
				return Headlines{}, stdio.GetEmptyErrorMessage()
			} else {
				_ = tx.Rollback()
				stdio.LogWarn("", "数据库SQL指令执行失败", err)
				return Headlines{}, stdio.GetErrorMessage(-500, "请求处理出错")
			}
		}
		item.Images = []string{image}
		if !expired && itemExpired < time.Now().Unix() {
			stdio.LogInfo("", "头条新闻数据过期")
			expired = true
		}
		items = append(items, item)
	}
	if items == nil {
		stdio.LogInfo("", "头条新闻详情不存在")
		return Headlines{}, stdio.GetEmptyErrorMessage()
	}
	var itemResult []NewsItem
	tx.Commit()
	for _, item := range items {
		localItem, errMessage := NewsManager.GetNewsById(item.Tid, item.Nid)
		if errMessage.HasInfo {
			return Headlines{}, stdio.GetEmptyErrorMessage()
		}
		item.Title = localItem.Title
		item.Summary = localItem.Summary
		itemResult = append(itemResult, item)
	}
	return Headlines{
		Exist:   true,
		Expired: expired,
		News:    itemResult,
	}, stdio.GetEmptyErrorMessage()
}

func (newsManagerImpl newsManagerImpl) UpdateHeadlines(headlines []NewsItem) stdio.MessagedError {
	tx, err := unit.Maria.Begin()
	if err != nil {
		stdio.LogWarn("", "数据库开始事务失败", err)
		return stdio.GetErrorMessage(-500, "请求处理出错")
	}
	//goland:noinspection SqlWithoutWhere
	state, err := tx.Prepare("delete from `news_headline`")
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn("", "数据库准备SQL指令失败", err)
		return stdio.GetErrorMessage(-500, "请求处理出错")
	}
	_, err = state.Exec()
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn("", "数据库准备SQL指令失败", err)
		return stdio.GetErrorMessage(-500, "请求处理出错")
	} else {
		tx.Commit()
	}
	for _, item := range headlines {
		tx, err := unit.Maria.Begin()
		if err != nil {
			stdio.LogWarn("", "数据库开始事务失败", err)
			return stdio.GetErrorMessage(-500, "请求处理出错")
		}
		state, err := tx.Prepare("insert into `news_headline` (`h_id`,`h_type_id`,`h_image`,`h_expired`) values (?, ?, ?, ?)")
		if err != nil {
			_ = tx.Rollback()
			stdio.LogWarn("", "数据库准备SQL指令失败", err)
			return stdio.GetErrorMessage(-500, "请求处理出错")
		}
		_, err = state.Exec(item.Nid, item.Tid, item.Images[0], time.Now().Unix()+86400)
		if err != nil {
			_ = tx.Rollback()
			stdio.LogWarn("", "数据库准备SQL指令失败", err)
			return stdio.GetErrorMessage(-500, "请求处理出错")
		} else {
			stdio.LogInfo("", "向数据库更新头条新闻成功")
			tx.Commit()
		}
	}
	return stdio.GetEmptyErrorMessage()
}

func (newsManagerImpl newsManagerImpl) CheckNewsExist(tid int, nid int) (bool, stdio.MessagedError) {
	tx, err := unit.Maria.Begin()
	if err != nil {
		stdio.LogWarn("", "数据库开始事务失败", err)
		return false, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	var state *sql.Stmt
	state, err = tx.Prepare("select `n_id` from `news` where `n_id`=? and `n_type_id`=?")
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn("", "数据库准备SQL指令失败", err)
		return false, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	rows := state.QueryRow(nid, tid)
	id := -1
	err = rows.Scan(&id)
	if err == nil {
		tx.Commit()
		return id != -1, stdio.GetEmptyErrorMessage()
	}
	if err == sql.ErrNoRows {
		tx.Commit()
		return false, stdio.GetEmptyErrorMessage()
	} else {
		_ = tx.Rollback()
		stdio.LogWarn("", "数据库SQL指令执行失败", err)
		return false, stdio.GetErrorMessage(-500, "请求处理出错")
	}
}

func (newsManagerImpl newsManagerImpl) GetTypeChart() ([]NewsTypeChartItem, stdio.MessagedError) {
	tx, err := unit.Maria.Begin()
	if err != nil {
		stdio.LogWarn("", "数据库开始事务失败", err)
		return nil, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	var state *sql.Stmt
	state, err = tx.Prepare("select * from `news_chart`")
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn("", "数据库准备SQL指令失败", err)
		return nil, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	rows, err := state.Query()
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn("", "数据库SQL指令执行失败", err)
		return nil, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	var charts []NewsTypeChartItem
	for rows.Next() {
		chart := NewsTypeChartItem{}
		err = rows.Scan(&chart.TypeId, &chart.TypeName, &chart.Out)
		if err != nil {
			_ = tx.Rollback()
			stdio.LogWarn("", "数据库SQL指令执行失败", err)
			return nil, stdio.GetErrorMessage(-500, "请求处理出错")
		}
		charts = append(charts, chart)
	}
	return charts, stdio.GetEmptyErrorMessage()
}

func (newsManagerImpl newsManagerImpl) UpdateTypeChart(chart []NewsTypeChartItem) stdio.MessagedError {
	for _, item := range chart {
		exist, errMessage := NewsManager.CheckChartExist(item.TypeId)
		if errMessage.HasInfo {
			return errMessage
		}
		tx, err := unit.Maria.Begin()
		if err != nil {
			stdio.LogWarn("", "数据库开始事务失败", err)
			return stdio.GetErrorMessage(-500, "请求处理出错")
		}
		var state *sql.Stmt
		if exist {
			state, err = tx.Prepare("update `news_chart` set `n_name`=? where `n_type_id`=?")
		} else {
			state, err = tx.Prepare("insert into `news_chart` (`n_name`, `n_type_id`) values (?, ?)")
		}
		if err != nil {
			stdio.LogWarn("", "数据库准备SQL指令失败", err)
			return stdio.GetErrorMessage(-500, "请求处理出错")
		}
		_, err = state.Exec(item.TypeName, item.TypeId)
		if err != nil {
			_ = tx.Rollback()
			stdio.LogWarn("", "数据库SQL指令执行失败", err)
			return stdio.GetErrorMessage(-500, "请求处理出错")
		}
		tx.Commit()
		if exist {
			stdio.LogInfo("", "向数据库插入新新闻类型字典成功")
		} else {
			stdio.LogInfo("", "向数据库更新新闻类型字典成功")
		}
	}
	return stdio.GetEmptyErrorMessage()
}

func (newsManagerImpl newsManagerImpl) CheckChartExist(nTypeId int) (bool, stdio.MessagedError) {
	tx, err := unit.Maria.Begin()
	if err != nil {
		stdio.LogWarn("", "数据库开始事务失败", err)
		return false, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	var state *sql.Stmt
	state, err = tx.Prepare("select `n_type_id` from `news_chart` where `n_type_id`=?")
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn("", "数据库准备SQL指令失败", err)
		return false, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	rows := state.QueryRow(nTypeId)
	id := ""
	err = rows.Scan(&id)
	if err == nil {
		tx.Commit()
		return id != "", stdio.GetEmptyErrorMessage()
	}
	if err == sql.ErrNoRows {
		tx.Commit()
		return false, stdio.GetEmptyErrorMessage()
	} else {
		_ = tx.Rollback()
		stdio.LogWarn("", "数据库SQL指令执行失败", err)
	}
	return false, stdio.GetErrorMessage(-500, "请求处理出错")
}
