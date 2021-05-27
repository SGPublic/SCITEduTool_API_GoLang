package manager

import (
	"SCITEduTool/Application/stdio"
	"SCITEduTool/Application/unit"
	"database/sql"
)

type chartManager interface {
	GetFacultyName(fId int) (string, stdio.MessagedError)
	GetSpecialtyName(fId int, sId int) (string, stdio.MessagedError)
	GetClassName(fId int, sId int, cId int) (string, stdio.MessagedError)
	GetChartIDWithClassName(cName string) (ChartIDItem, stdio.MessagedError)
	WriteFacultyName(fId int, fName string) stdio.MessagedError
	WriteSpecialtyName(fId int, sId int, sName string) stdio.MessagedError
	WriteClassName(fId int, sId int, cId int, cName string) stdio.MessagedError
}

type chartManagerImpl struct{}

var ChartManager chartManager = chartManagerImpl{}

func (chartManagerImpl chartManagerImpl) GetFacultyName(fId int) (string, stdio.MessagedError) {
	tx, err := unit.Maria.Begin()
	if err != nil {
		stdio.LogWarn("", "数据库开始事务失败", err)
		return "", stdio.GetErrorMessage(-500, "请求处理出错")
	}
	state, err := tx.Prepare("select `f_name` from `faculty_chart` where `f_id`=?")
	if err != nil {
		stdio.LogWarn("", "数据库准备SQL指令失败", err)
		return "", stdio.GetErrorMessage(-500, "请求处理出错")
	}
	rows := state.QueryRow(fId)
	var fName string
	err = rows.Scan(&fName)
	if err == nil {
		tx.Commit()
		return fName, stdio.GetEmptyErrorMessage()
	}
	_ = tx.Rollback()
	if err == sql.ErrNoRows {
		return "", stdio.GetEmptyErrorMessage()
	} else {
		stdio.LogWarn("", "数据库SQL指令执行失败", err)
		return "", stdio.GetErrorMessage(-500, "请求处理出错")
	}
}

func (chartManagerImpl chartManagerImpl) GetSpecialtyName(fId int, sId int) (string, stdio.MessagedError) {
	tx, err := unit.Maria.Begin()
	if err != nil {
		stdio.LogWarn("", "数据库开始事务失败", err)
		return "", stdio.GetErrorMessage(-500, "请求处理出错")
	}
	state, err := tx.Prepare("select `s_name` from `specialty_chart` where `f_id`=? and `s_id`=?")
	if err != nil {
		stdio.LogWarn("", "数据库准备SQL指令失败", err)
		return "", stdio.GetErrorMessage(-500, "请求处理出错")
	}
	rows := state.QueryRow(fId, sId)
	var sName string
	err = rows.Scan(&sName)
	//err := rows.Scan(&item)
	if err == nil {
		tx.Commit()
		return sName, stdio.GetEmptyErrorMessage()
	}
	_ = tx.Rollback()
	if err == sql.ErrNoRows {
		return "", stdio.GetEmptyErrorMessage()
	} else {
		stdio.LogWarn("", "数据库SQL指令执行失败", err)
		return "", stdio.GetErrorMessage(-500, "请求处理出错")
	}
}

func (chartManagerImpl chartManagerImpl) GetClassName(fId int, sId int, cId int) (string, stdio.MessagedError) {
	tx, err := unit.Maria.Begin()
	if err != nil {
		stdio.LogWarn("", "数据库开始事务失败", err)
		return "", stdio.GetErrorMessage(-500, "请求处理出错")
	}
	state, err := tx.Prepare("select `c_name` from `class_chart` where `f_id`=? and `s_id`=? and `c_id`=?")
	if err != nil {
		stdio.LogWarn("", "数据库准备SQL指令失败", err)
		return "", stdio.GetErrorMessage(-500, "请求处理出错")
	}
	rows := state.QueryRow(fId, sId, cId)
	var cName string
	err = rows.Scan(&cName)
	if err == nil {
		tx.Commit()
		return cName, stdio.GetEmptyErrorMessage()
	}
	_ = tx.Rollback()
	if err == sql.ErrNoRows {
		return "", stdio.GetEmptyErrorMessage()
	} else {
		stdio.LogWarn("", "数据库SQL指令执行失败", err)
		return "", stdio.GetErrorMessage(-500, "请求处理出错")
	}
}

type ChartIDItem struct {
	Exist       bool
	FacultyId   int
	SpecialtyId int
}

func (chartManagerImpl chartManagerImpl) GetChartIDWithClassName(cName string) (ChartIDItem, stdio.MessagedError) {
	tx, err := unit.Maria.Begin()
	if err != nil {
		stdio.LogWarn("", "数据库开始事务失败", err)
		return ChartIDItem{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	state, err := tx.Prepare("select `f_id`,`s_id` from `class_chart` where `c_name`=?")
	if err != nil {
		stdio.LogWarn("", "数据库准备SQL指令失败", err)
		return ChartIDItem{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	item := ChartIDItem{}
	rows := state.QueryRow(cName)
	err = rows.Scan(&item.FacultyId, &item.SpecialtyId)
	if err == nil {
		tx.Commit()
		return item, stdio.GetEmptyErrorMessage()
	}
	_ = tx.Rollback()
	if err == sql.ErrNoRows {
		return ChartIDItem{}, stdio.GetEmptyErrorMessage()
	} else {
		stdio.LogWarn("", "数据库SQL指令执行失败", err)
		return ChartIDItem{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
}

func (chartManagerImpl chartManagerImpl) WriteFacultyName(fId int, fName string) stdio.MessagedError {
	fNameExist, errMessage := ChartManager.GetFacultyName(fId)
	if errMessage.HasInfo {
		return errMessage
	}
	tx, err := unit.Maria.Begin()
	if err != nil {
		stdio.LogWarn("", "数据库开始事务失败", err)
		return stdio.GetErrorMessage(-500, "请求处理出错")
	}
	var state *sql.Stmt
	if fNameExist == "" {
		state, err = tx.Prepare("insert into `faculty_chart` (`f_id`, `f_name`) values (?, ?)")
	} else {
		state, err = tx.Prepare("update `faculty_chart` set `f_name`=? where `f_id`=?")
	}
	if err != nil {
		stdio.LogWarn("", "数据库准备SQL指令失败", err)
		return stdio.GetErrorMessage(-500, "请求处理出错")
	}
	if fNameExist == "" {
		_, err = state.Exec(fId, fName)
	} else {
		_, err = state.Exec(fName, fId)
	}
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn("", "数据库SQL指令执行失败", err)
		return stdio.GetErrorMessage(-500, "请求处理出错")
	}
	tx.Commit()
	if fNameExist == "" {
		stdio.LogVerbose("", "向数据库插入新学院名称字典成功")
	} else {
		stdio.LogVerbose("", "向数据库更新学院名称字典成功")
	}
	return stdio.GetEmptyErrorMessage()
}

func (chartManagerImpl chartManagerImpl) WriteSpecialtyName(fId int, sId int, sName string) stdio.MessagedError {
	sNameExist, errMessage := ChartManager.GetSpecialtyName(fId, sId)
	if errMessage.HasInfo {
		return errMessage
	}
	tx, err := unit.Maria.Begin()
	if err != nil {
		stdio.LogWarn("", "数据库开始事务失败", err)
		return stdio.GetErrorMessage(-500, "请求处理出错")
	}
	var state *sql.Stmt
	if sNameExist == "" {
		state, err = tx.Prepare("insert into `specialty_chart` (`f_id`, `s_id`, `s_name`) values (?, ?, ?)")
	} else {
		state, err = tx.Prepare("update `specialty_chart` set `s_name`=? where `f_id`=? and `s_id`=?")
	}
	if err != nil {
		stdio.LogWarn("", "数据库准备SQL指令失败", err)
		return stdio.GetErrorMessage(-500, "请求处理出错")
	}
	if sNameExist == "" {
		_, err = state.Exec(fId, sId, sName)
	} else {
		_, err = state.Exec(sName, fId, sId)
	}
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn("", "数据库SQL指令执行失败", err)
		return stdio.GetErrorMessage(-500, "请求处理出错")
	}
	tx.Commit()
	if sNameExist == "" {
		stdio.LogVerbose("", "向数据库插入新专业名称字典成功")
	} else {
		stdio.LogVerbose("", "向数据库更新专业名称字典成功")
	}
	return stdio.GetEmptyErrorMessage()
}

func (chartManagerImpl chartManagerImpl) WriteClassName(fId int, sId int, cId int, cName string) stdio.MessagedError {
	cNameExist, errMessage := ChartManager.GetClassName(fId, sId, cId)
	if errMessage.HasInfo {
		return errMessage
	}
	tx, err := unit.Maria.Begin()
	if err != nil {
		stdio.LogWarn("", "数据库开始事务失败", err)
		return stdio.GetErrorMessage(-500, "请求处理出错")
	}
	var state *sql.Stmt
	if cNameExist == "" {
		state, err = tx.Prepare("insert into `class_chart` (`f_id`, `s_id`, `c_id`, `c_name`) values (?, ?, ?, ?)")
	} else {
		state, err = tx.Prepare("update `class_chart` set `c_name`=? where `f_id`=? and `s_id`=? and `c_id`=?")
	}
	if err != nil {
		stdio.LogWarn("", "数据库准备SQL指令失败", err)
		return stdio.GetErrorMessage(-500, "请求处理出错")
	}
	if cNameExist == "" {
		_, err = state.Exec(fId, sId, cId, cName)
	} else {
		_, err = state.Exec(cName, fId, sId, cId)
	}
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn("", "数据库SQL指令执行失败", err)
		return stdio.GetErrorMessage(-500, "请求处理出错")
	}
	tx.Commit()
	if cNameExist == "" {
		stdio.LogVerbose("", "向数据库插入新班级名称字典成功")
	} else {
		stdio.LogVerbose("", "向数据库更新班级名称字典成功")
	}
	return stdio.GetEmptyErrorMessage()
}
