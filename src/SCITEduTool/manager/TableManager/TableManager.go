package TableManager

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"SCITEduTool/manager/InfoManager"
	"SCITEduTool/unit/SQLStaticUnit"
	"SCITEduTool/unit/StdOutUnit"
)

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
	Object [6][5]LessonItem `json:"table"`
}

type TableContent struct {
	Exist   bool
	Expired bool
	Table   string
}

func Get(username string, info InfoManager.UserInfo, year string, semester int) (TableContent, StdOutUnit.MessagedError) {
	tableId := GetTableId(info.Specialty, info.Grade, info.Class, year, semester)
	tx, err := SQLStaticUnit.Maria.Begin()
	if err != nil {
		StdOutUnit.Warn(username, "数据库开始事务失败", err)
		return TableContent{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	state, err := tx.Prepare("select `t_content`,`t_expired` from `class_schedule` where `t_id`=?")
	if err != nil {
		_ = tx.Rollback()
		StdOutUnit.Warn(username, "数据库准备SQL指令失败", err)
		return TableContent{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	rows := state.QueryRow(tableId)
	table := TableContent{}
	var expired int64
	err = rows.Scan(&table.Table, &expired)
	if err == nil {
		tx.Commit()
		table.Exist = true
		table.Expired = expired < time.Now().Unix()
		return table, StdOutUnit.GetEmptyErrorMessage()
	}
	if err == sql.ErrNoRows {
		tx.Commit()
		return TableContent{}, StdOutUnit.GetEmptyErrorMessage()
	} else {
		_ = tx.Rollback()
		StdOutUnit.Warn(username, "数据库SQL指令执行失败", err)
		return TableContent{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
}

func Update(username string, info InfoManager.UserInfo, year string, semester int, table TableObject) StdOutUnit.MessagedError {
	tableId := GetTableId(info.Specialty, info.Grade, info.Class, year, semester)
	tableContent, err := json.Marshal(table)
	if err != nil {
		StdOutUnit.Warn(username, "数据库开始事务失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	tableString := string(tableContent)

	exist, errMessage := CheckTableExist(username, tableId)
	if errMessage.HasInfo {
		return errMessage
	}
	tx, err := SQLStaticUnit.Maria.Begin()
	if err != nil {
		StdOutUnit.Warn(username, "数据库开始事务失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	var state *sql.Stmt
	if !exist {
		state, err = tx.Prepare("insert into `class_schedule` (`t_id`, `t_faculty` ,`t_specialty`, `t_class`, `t_grade`, `t_school_year`, `t_semester`, `t_content`, `t_expired`) values (?, ?, ?, ?, ?, ?, ?, ?, ?)")
	} else {
		state, err = tx.Prepare("update `class_schedule` set `t_content`=?, `t_expired`=? where `t_id`=?")
	}
	if err != nil {
		_ = tx.Rollback()
		StdOutUnit.Warn(username, "数据库准备SQL指令失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	if !exist {
		_, err = state.Exec(tableId, info.Faculty, info.Specialty, info.Class, info.Grade, year, semester, tableString,
			time.Now().Unix()+1296000)
	} else {
		_, err = state.Exec(tableId, tableString, time.Now().Unix()+1296000, username)
	}
	if err != nil {
		_ = tx.Rollback()
		StdOutUnit.Warn(username, "数据库SQL指令执行失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	tx.Commit()
	if !exist {
		StdOutUnit.Verbose(username, "向数据库插入新课表数据成功")
	} else {
		StdOutUnit.Verbose(username, "向数据库更新课表数据成功")
	}
	return StdOutUnit.GetEmptyErrorMessage()
}

func GetTableId(specialty int, grade int, class int, year string, semester int) string {
	classId := "0" + strconv.Itoa(class)
	classId = classId[len(classId)-2:]
	return strconv.Itoa(grade) + strconv.Itoa(specialty) + year + strconv.Itoa(semester) +
		strconv.Itoa(grade)[2:] + strconv.Itoa(specialty) + classId
}

func CheckTableExist(username string, tableId string) (bool, StdOutUnit.MessagedError) {
	tx, err := SQLStaticUnit.Maria.Begin()
	if err != nil {
		StdOutUnit.Warn(username, "数据库开始事务失败", err)
		return false, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	var state *sql.Stmt
	state, err = tx.Prepare("select `t_id` from `class_schedule` where `t_id`=?")
	if err != nil {
		_ = tx.Rollback()
		StdOutUnit.Warn(username, "数据库准备SQL指令失败", err)
		return false, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	rows := state.QueryRow(tableId)
	id := ""
	err = rows.Scan(&id)
	if err == nil {
		tx.Commit()
		return id != "", StdOutUnit.GetEmptyErrorMessage()
	}
	if err == sql.ErrNoRows {
		tx.Commit()
		return false, StdOutUnit.GetEmptyErrorMessage()
	} else {
		_ = tx.Rollback()
		StdOutUnit.Warn(username, "数据库SQL指令执行失败", err)
	}
	return false, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
}
