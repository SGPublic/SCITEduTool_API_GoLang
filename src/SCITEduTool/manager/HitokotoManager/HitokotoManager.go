package HitokotoManager

import (
	"database/sql"
	"time"

	"SCITEduTool/unit/SQLStaticUnit"
	"SCITEduTool/unit/StdOutUnit"
)

type HitokotoItem struct {
	Exist      bool
	Index      int    `json:"id"`
	Content    string `json:"hitokoto"`
	Type       string `json:"type"`
	From       string `json:"from"`
	FromWho    string `json:"from_who"`
	Creator    string `json:"creator"`
	CreatorUid int    `json:"creator_uid"`
	Reviewer   int    `json:"reviewer"`
	InsertAt   string `json:"created_at"`
	Length     int    `json:"length"`
}

func Get() (HitokotoItem, StdOutUnit.MessagedError) {
	hitokoto := HitokotoItem{}
	tx, err := SQLStaticUnit.Maria.Begin()
	if err != nil {
		StdOutUnit.Warn("", "数据库开始事务失败", err)
		return HitokotoItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	state, err := tx.Prepare("select `h_id` from `hitokoto` where `h_insert_at`>?")
	if err != nil {
		_ = tx.Rollback()
		StdOutUnit.Warn("", "数据库准备SQL指令失败", err)
		return HitokotoItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	rows := state.QueryRow(time.Now().Unix() - 3600)
	hId := 0
	err = rows.Scan(&hId)
	if err == nil {
		goto rand
	}
	if err == sql.ErrNoRows {
		tx.Commit()
		return HitokotoItem{}, StdOutUnit.GetEmptyErrorMessage()
	} else {
		_ = tx.Rollback()
		StdOutUnit.Warn("", "数据库SQL指令执行失败", err)
		return HitokotoItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

rand:
	tx, err = SQLStaticUnit.Maria.Begin()
	if err != nil {
		StdOutUnit.Warn("", "数据库开始事务失败", err)
		return HitokotoItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	state, err = tx.Prepare("select `h_content`,`h_from`,`h_length` from `hitokoto` where `h_id` >= (select floor(RAND() * (select MAX(`h_id`) from `hitokoto`))) order by `h_id` limit 1")
	if err != nil {
		_ = tx.Rollback()
		StdOutUnit.Warn("", "数据库准备SQL指令失败", err)
		return HitokotoItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	rows = state.QueryRow()
	err = rows.Scan(&hitokoto.Content, &hitokoto.From, &hitokoto.Length)
	if err == nil {
		tx.Commit()
		hitokoto.Exist = true
		return hitokoto, StdOutUnit.GetEmptyErrorMessage()
	}
	if err == sql.ErrNoRows {
		tx.Commit()
		return HitokotoItem{}, StdOutUnit.GetEmptyErrorMessage()
	} else {
		_ = tx.Rollback()
		StdOutUnit.Warn("", "数据库SQL指令执行失败", err)
		return HitokotoItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
}

func Insert(item HitokotoItem) StdOutUnit.MessagedError {
	exit, errMessage := CheckHitokotoExist(item.Index)
	if errMessage.HasInfo {
		return errMessage
	}
	if exit {
		return StdOutUnit.GetEmptyErrorMessage()
	}

	tx, err := SQLStaticUnit.Maria.Begin()
	if err != nil {
		StdOutUnit.Warn("", "数据库开始事务失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	state, err := tx.Prepare("insert into `hitokoto` (h_index, h_content, h_type, h_from, h_from_who, h_creator, h_creator_uid, h_reviewer, h_insert_at, h_length) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		_ = tx.Rollback()
		StdOutUnit.Warn("", "数据库准备SQL指令失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	_, err = state.Exec(item.Index, item.Content, item.Type, item.From, item.FromWho, item.Creator,
		item.CreatorUid, item.Reviewer, time.Now().Unix(), item.Length)
	if err == nil {
		tx.Commit()
		StdOutUnit.Verbose("", "向数据库插入新Hitokoto成功")
		return StdOutUnit.GetEmptyErrorMessage()
	} else {
		_ = tx.Rollback()
		StdOutUnit.Warn("", "数据库查询失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
}

func CheckHitokotoExist(index int) (bool, StdOutUnit.MessagedError) {
	tx, err := SQLStaticUnit.Maria.Begin()
	if err != nil {
		StdOutUnit.Warn("", "数据库开始事务失败", err)
		return false, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	var state *sql.Stmt
	state, err = tx.Prepare("select `h_index` from `hitokoto` where `h_index`=?")
	if err != nil {
		_ = tx.Rollback()
		StdOutUnit.Warn("", "数据库准备SQL指令失败", err)
		return false, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	rows := state.QueryRow(index)
	id := -1
	err = rows.Scan(&id)
	if err == nil {
		tx.Commit()
		return id != -1, StdOutUnit.GetEmptyErrorMessage()
	}
	if err == sql.ErrNoRows {
		tx.Commit()
		return false, StdOutUnit.GetEmptyErrorMessage()
	} else {
		_ = tx.Rollback()
		StdOutUnit.Warn("", "数据库SQL指令执行失败", err)
		return false, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
}
