package manager

import (
	"SCITEduTool/Application/stdio"
	"SCITEduTool/Application/unit"
	"database/sql"
	"encoding/json"
	"time"
)

type tableManager interface {
	Get(username string, info UserInfo, year string, semester int) (TableContent, stdio.MessagedError)
	Update(username string, info UserInfo, year string, semester int, tableId string, table TableObject) stdio.MessagedError
	CheckTableExist(username string, tableId string) (bool, stdio.MessagedError)
}

type tableManagerImpl struct{}

var TableManager tableManager = tableManagerImpl{}

type LessonSingleItem struct {
	Name    string `json:"name"`
	Range   []int  `json:"range"`
	Teacher string `json:"teacher"`
	Room    string `json:"room"`
}

type LessonItem struct {
	Data []LessonSingleItem `json:"data"`
}

type TableObject struct {
	Object [7][5]LessonItem `json:"table"`
}

type TableContent struct {
	Exist   bool
	Expired bool
	Table   string
}

func (tableManagerImpl tableManagerImpl) Get(username string, info UserInfo, year string, semester int) (TableContent, stdio.MessagedError) {
	tx, err := unit.Maria.Begin()
	if err != nil {
		stdio.LogWarn(username, "数据库开始事务失败", err)
		return TableContent{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	state, err := tx.Prepare("select `t_content`,`t_expired` from `class_schedule` where `t_faculty`=? and `t_specialty`=? and `t_class`=? and `t_school_year`=? and `t_semester`=?")
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn(username, "数据库准备SQL指令失败", err)
		return TableContent{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	rows := state.QueryRow(info.Faculty, info.Specialty, info.Class, year, semester)
	table := TableContent{}
	var expired int64
	err = rows.Scan(&table.Table, &expired)
	if err == nil {
		tx.Commit()
		table.Exist = true
		table.Expired = expired < time.Now().Unix()
		return table, stdio.GetEmptyErrorMessage()
	}
	if err == sql.ErrNoRows {
		tx.Commit()
		return TableContent{}, stdio.GetEmptyErrorMessage()
	} else {
		_ = tx.Rollback()
		stdio.LogWarn(username, "数据库SQL指令执行失败", err)
		return TableContent{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
}

func (tableManagerImpl tableManagerImpl) Update(username string, info UserInfo, year string, semester int, tableId string, table TableObject) stdio.MessagedError {
	tableContent, err := json.Marshal(table)
	if err != nil {
		stdio.LogWarn(username, "数据库开始事务失败", err)
		return stdio.GetErrorMessage(-500, "请求处理出错")
	}
	tableString := string(tableContent)

	exist, errMessage := TableManager.CheckTableExist(username, tableId)
	if errMessage.HasInfo {
		return errMessage
	}
	tx, err := unit.Maria.Begin()
	if err != nil {
		stdio.LogWarn(username, "数据库开始事务失败", err)
		return stdio.GetErrorMessage(-500, "请求处理出错")
	}
	var state *sql.Stmt
	if !exist {
		state, err = tx.Prepare("insert into `class_schedule` (`t_id`, `t_faculty` ,`t_specialty`, `t_class`, `t_grade`, `t_school_year`, `t_semester`, `t_content`, `t_expired`) values (?, ?, ?, ?, ?, ?, ?, ?, ?)")
	} else {
		state, err = tx.Prepare("update `class_schedule` set `t_content`=?, `t_expired`=? where `t_id`=?")
	}
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn(username, "数据库准备SQL指令失败", err)
		return stdio.GetErrorMessage(-500, "请求处理出错")
	}
	if !exist {
		_, err = state.Exec(tableId, info.Faculty, info.Specialty, info.Class, info.Grade, year, semester, tableString,
			time.Now().Unix()+1296000)
	} else {
		_, err = state.Exec(tableString, time.Now().Unix()+1296000, tableId)
	}
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn(username, "数据库SQL指令执行失败", err)
		return stdio.GetErrorMessage(-500, "请求处理出错")
	}
	tx.Commit()
	if !exist {
		stdio.LogVerbose(username, "向数据库插入新课表数据成功")
	} else {
		stdio.LogVerbose(username, "向数据库更新课表数据成功")
	}
	return stdio.GetEmptyErrorMessage()
}

func (tableManagerImpl tableManagerImpl) CheckTableExist(username string, tableId string) (bool, stdio.MessagedError) {
	tx, err := unit.Maria.Begin()
	if err != nil {
		stdio.LogWarn(username, "数据库开始事务失败", err)
		return false, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	var state *sql.Stmt
	state, err = tx.Prepare("select `t_id` from `class_schedule` where `t_id`=?")
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn(username, "数据库准备SQL指令失败", err)
		return false, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	rows := state.QueryRow(tableId)
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
		stdio.LogWarn(username, "数据库SQL指令执行失败", err)
	}
	return false, stdio.GetErrorMessage(-500, "请求处理出错")
}
