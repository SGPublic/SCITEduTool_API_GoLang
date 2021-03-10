package SessionManager

import (
	"SCITEduTool/unit/SQLStaticUnit"
	"SCITEduTool/unit/StdOutUnit"
	"database/sql"
	"time"
)

type userSessionItem struct {
	Session   string
	Expired   int64
	Effective int
}

type SessionItem struct {
	Session   string
	Identify  int
	Exist     bool
	Expired   bool
	Effective bool
}

func Get(username string) (SessionItem, StdOutUnit.MessagedError) {
	tx, err := SQLStaticUnit.Maria.Begin()
	if err != nil {
		StdOutUnit.Warn.String(username, "数据库开始事务失败", err)
		return SessionItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	state, err := tx.Prepare("select `u_session`,`u_session_expired`,`u_token_effective` from `user_token` where `u_id`=?")
	if err != nil {
		_ = tx.Rollback()
		StdOutUnit.Warn.String(username, "数据库准备SQL指令失败", err)
		return SessionItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	rows := state.QueryRow(username)
	var item userSessionItem
	err = rows.Scan(&item.Session, &item.Expired, &item.Effective)
	//err := rows.Scan(&item)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Commit()
			StdOutUnit.Info.String(username, "用户不存在")
			return SessionItem{}, StdOutUnit.GetEmptyErrorMessage()
		} else {
			_ = tx.Rollback()
			StdOutUnit.Warn.String(username, "数据库SQL指令执行失败", err)
			return SessionItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
		}
	}
	tx.Commit()
	tx, err = SQLStaticUnit.Maria.Begin()
	if err != nil {
		StdOutUnit.Warn.String(username, "数据库开始事务失败", err)
		return SessionItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	state, err = tx.Prepare("select `u_identify` from `user_info` where `u_id`=?")
	if err != nil {
		_ = tx.Rollback()
		StdOutUnit.Warn.String(username, "数据库准备SQL指令失败", err)
		return SessionItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	rows = state.QueryRow(username)
	identify := -1
	err = rows.Scan(&identify)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Commit()
			StdOutUnit.Info.String(username, "用户身份未知")
			return SessionItem{}, StdOutUnit.GetEmptyErrorMessage()
		} else {
			_ = tx.Rollback()
			StdOutUnit.Warn.String(username, "数据库SQL指令执行失败", err)
			return SessionItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
		}
	}
	tx.Commit()
	if identify < 0 {
		StdOutUnit.Warn.String(username, "用户身份获取失败", nil)
		return SessionItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	if item.Session != "" {
		return SessionItem{
			Session:   item.Session,
			Identify:  identify,
			Exist:     true,
			Effective: item.Effective == 1,
			Expired:   item.Expired < time.Now().Unix(),
		}, StdOutUnit.GetEmptyErrorMessage()
	} else {
		return SessionItem{
			Exist: false,
		}, StdOutUnit.GetEmptyErrorMessage()
	}
}

func Update(username string, password string, session string, identify int) StdOutUnit.MessagedError {
	exist, errMessage := CheckUserExist(username, "user_token")
	if errMessage.HasInfo {
		return errMessage
	}
	tx, err := SQLStaticUnit.Maria.Begin()
	if err != nil {
		StdOutUnit.Warn.String(username, "数据库开始事务失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	var state *sql.Stmt
	if !exist {
		state, err = tx.Prepare("insert into `user_token` (`u_id`, `u_password` ,`u_session`, `u_session_expired`, `u_token_effective`) values (?, ?, ?, ?, 1)")
	} else {
		state, err = tx.Prepare("update `user_token` set `u_session`=?, `u_session_expired`=?, `u_token_effective`=1, `u_password`=? where `u_id`=?")
	}
	if err != nil {
		StdOutUnit.Warn.String(username, "数据库准备SQL指令失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	if !exist {
		_, err = state.Exec(username, password, session, time.Now().Unix()+1800)
	} else {
		_, err = state.Exec(session, time.Now().Unix()+1800, password, username)
	}
	if err == nil {
		tx.Commit()
		if !exist {
			StdOutUnit.Verbose.String(username, "向数据库插入新 ASP.NET_SessionId 成功")
		} else {
			StdOutUnit.Verbose.String(username, "向数据库更新 ASP.NET_SessionId 成功")
		}
	} else {
		_ = tx.Rollback()
		StdOutUnit.Warn.String(username, "数据库SQL指令执行失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	exist, errMessage = CheckUserExist(username, "user_info")
	if errMessage.HasInfo {
		return errMessage
	}
	tx, err = SQLStaticUnit.Maria.Begin()
	if err != nil {
		StdOutUnit.Warn.String(username, "数据库开始事务失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	if !exist {
		state, err = tx.Prepare("insert into `user_info` (`u_id`, `u_identify`, `u_info_expired`) values (?, ?, 0)")
	} else {
		state, err = tx.Prepare("update `user_info` set `u_identify`=? where `u_id`=?")
	}
	if err != nil {
		_ = tx.Rollback()
		StdOutUnit.Warn.String(username, "数据库准备SQL指令失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	if !exist {
		_, err = state.Exec(username, identify)
	} else {
		_, err = state.Exec(identify, username)
	}
	if err == nil {
		tx.Commit()
		if !exist {
			StdOutUnit.Verbose.String(username, "向数据库插入新用户身份成功")
		} else {
			StdOutUnit.Verbose.String(username, "向数据库更新用户身份成功")
		}
		return StdOutUnit.GetEmptyErrorMessage()
	} else {
		_ = tx.Rollback()
		StdOutUnit.Warn.String(username, "数据库查询失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
}

func GetUserPassword(username string, password string) (string, StdOutUnit.MessagedError) {
	pass := password
	if pass == "" {
		tx, err := SQLStaticUnit.Maria.Begin()
		if err != nil {
			StdOutUnit.Warn.String(username, "数据库开始事务失败", err)
			return "", StdOutUnit.GetErrorMessage(-500, "请求处理出错")
		}
		var state *sql.Stmt
		state, err = tx.Prepare("select `u_password` from `user_token` where `u_id`=?")
		if err != nil {
			_ = tx.Rollback()
			StdOutUnit.Warn.String(username, "数据库准备SQL指令失败", err)
			return "", StdOutUnit.GetErrorMessage(-500, "请求处理出错")
		}
		rows := state.QueryRow(username)
		err = rows.Scan(&pass)
		if err != nil {
			if err == sql.ErrNoRows {
				tx.Commit()
				return "", StdOutUnit.GetEmptyErrorMessage()
			} else {
				_ = tx.Rollback()
				StdOutUnit.Warn.String(username, "数据库SQL指令执行失败", err)
				return "", StdOutUnit.GetErrorMessage(-500, "请求处理出错")
			}
		}
		tx.Commit()
		if pass == "" {
			return "", StdOutUnit.GetEmptyErrorMessage()
		}
	}
	return pass, StdOutUnit.GetEmptyErrorMessage()
}

func CheckUserExist(username string, table string) (bool, StdOutUnit.MessagedError) {
	tx, err := SQLStaticUnit.Maria.Begin()
	if err != nil {
		StdOutUnit.Warn.String(username, "数据库开始事务失败", err)
		return false, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	var state *sql.Stmt
	state, err = tx.Prepare("select `u_id` from `" + table + "` where `u_id`=?")
	if err != nil {
		_ = tx.Rollback()
		StdOutUnit.Warn.String(username, "数据库准备SQL指令失败", err)
		return false, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	rows := state.QueryRow(username)
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
		StdOutUnit.Warn.String(username, "数据库SQL指令执行失败", err)
	}
	return false, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
}
