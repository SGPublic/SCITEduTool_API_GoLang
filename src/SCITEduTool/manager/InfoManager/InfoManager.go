package InfoManager

import (
	"SCITEduTool/manager/SessionManager"
	"SCITEduTool/unit/SQLStaticUnit"
	"SCITEduTool/unit/StdOutUnit"
	"database/sql"
	"time"
)

type ChartItem struct {
	Name string
	ID   int
}

type UserInfo struct {
	Exist     bool
	Expired   bool
	Name      string
	Identify  int
	Grade     int
	Faculty   int
	Specialty int
	Class     int
}

func Get(username string) (UserInfo, StdOutUnit.MessagedError) {
	exist, errMessage := SessionManager.CheckUserExist(username, "user_info")
	if errMessage.HasInfo {
		return UserInfo{}, errMessage
	}
	if !exist {
		return UserInfo{}, StdOutUnit.GetEmptyErrorMessage()
	}
	tx, err := SQLStaticUnit.Maria.Begin()
	if err != nil {
		StdOutUnit.Warn.String(username, err.Error())
		return UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	state, err := tx.Prepare("select `u_name`,`u_identify`,`u_faculty`,`u_specialty`,`u_class`,`u_info_expired` from `user_info` where `u_id`=?")
	if err != nil {
		StdOutUnit.Warn.String(username, err.Error())
		return UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	rows := state.QueryRow(username)
	info := UserInfo{}
	var expired int64
	err = rows.Scan(&info.Name, &info.Identify, &info.Faculty, &info.Specialty, &info.Class, &expired)
	if err != nil {
		if err == sql.ErrNoRows {
			return UserInfo{}, StdOutUnit.GetEmptyErrorMessage()
		} else {
			_ = tx.Rollback()
			StdOutUnit.Warn.String(username, err.Error())
			return UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
		}
	}
	tx.Commit()
	info.Exist = true
	info.Expired = expired < time.Now().Unix()
	return info, StdOutUnit.GetEmptyErrorMessage()
}

func Update(username string, name string, faculty int, specialty int, class int, grade int) StdOutUnit.MessagedError {
	exist, errMessage := SessionManager.CheckUserExist(username, "user_info")
	if errMessage.HasInfo {
		return errMessage
	}
	tx, err := SQLStaticUnit.Maria.Begin()
	if err != nil {
		StdOutUnit.Warn.String(username, err.Error())
		return StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	var state *sql.Stmt
	if !exist {
		state, err = tx.Prepare("insert into `user_info` (`u_id`, `u_name` ,`u_faculty`, `u_specialty`, `u_class`, `u_grade`, `u_info_expired`) values (?, ?, ?, ?, ?, ?, ?)")
	} else {
		state, err = tx.Prepare("update `user_info` set `u_name`=?, `u_faculty`=?, `u_specialty`=?, `u_class`=?, `u_grade`=?, `u_info_expired`=? where `u_id`=?")
	}
	if err != nil {
		StdOutUnit.Warn.String(username, err.Error())
		return StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	if !exist {
		_, err = state.Exec(username, name, faculty, specialty, class, grade, time.Now().Unix()+1296000)
	} else {
		_, err = state.Exec(name, faculty, specialty, class, grade, time.Now().Unix()+1296000, username)
	}
	if err != nil {
		_ = tx.Rollback()
		StdOutUnit.Warn.String(username, err.Error())
		return StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	tx.Commit()
	if !exist {
		StdOutUnit.Verbose.String(username, "向数据库插入新 ASP.NET_SessionId 成功")
	} else {
		StdOutUnit.Verbose.String(username, "向数据库更新 ASP.NET_SessionId 成功")
	}
	return StdOutUnit.GetEmptyErrorMessage()
}

func SetUserInfoExpired(username string) StdOutUnit.MessagedError {
	exist, errMessage := SessionManager.CheckUserExist(username, "user_info")
	if errMessage.HasInfo {
		return errMessage
	}
	if !exist {
		StdOutUnit.Warn.String(username, "用户信息不存在")
		return StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	tx, err := SQLStaticUnit.Maria.Begin()
	if err != nil {
		StdOutUnit.Warn.String(username, err.Error())
		return StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	state, err := tx.Prepare("update `user_token` set `u_token_effective`=0 where `u_id`=?")
	if err != nil {
		StdOutUnit.Warn.String(username, err.Error())
		return StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	_, err = state.Exec(username)
	if err == nil {
		tx.Commit()
		StdOutUnit.Verbose.String(username, "标记用户token失效成功")
		return StdOutUnit.GetEmptyErrorMessage()
	} else {
		_ = tx.Rollback()
		StdOutUnit.Warn.String(username, err.Error())
		return StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
}
