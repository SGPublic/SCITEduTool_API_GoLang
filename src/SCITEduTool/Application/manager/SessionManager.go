package manager

import (
	"SCITEduTool/Application/stdio"
	"SCITEduTool/Application/unit"
	"database/sql"
	"time"
)

type sessionManager interface {
	Get(username string) (SessionItem, stdio.MessagedError)
	Update(username string, password string, session string, identify int) stdio.MessagedError
	GetUserPassword(username string, password string) (string, stdio.MessagedError)
	CheckUserExist(username string, table string) (bool, stdio.MessagedError)
}

type sessionManagerImpl struct{}

var SessionManager sessionManager = sessionManagerImpl{}

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

func (sessionManagerImpl sessionManagerImpl) Get(username string) (SessionItem, stdio.MessagedError) {
	tx, err := unit.Maria.Begin()
	if err != nil {
		stdio.LogWarn(username, "数据库开始事务失败", err)
		return SessionItem{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	state, err := tx.Prepare("select `u_session`,`u_session_expired`,`u_token_effective` from `user_token` where `u_id`=?")
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn(username, "数据库准备SQL指令失败", err)
		return SessionItem{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	rows := state.QueryRow(username)
	var item userSessionItem
	err = rows.Scan(&item.Session, &item.Expired, &item.Effective)
	//err := rows.Scan(&item)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Commit()
			stdio.LogInfo(username, "用户不存在")
			return SessionItem{}, stdio.GetEmptyErrorMessage()
		} else {
			_ = tx.Rollback()
			stdio.LogWarn(username, "数据库SQL指令执行失败", err)
			return SessionItem{}, stdio.GetErrorMessage(-500, "请求处理出错")
		}
	}
	tx.Commit()
	tx, err = unit.Maria.Begin()
	if err != nil {
		stdio.LogWarn(username, "数据库开始事务失败", err)
		return SessionItem{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	state, err = tx.Prepare("select `u_identify` from `user_info` where `u_id`=?")
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn(username, "数据库准备SQL指令失败", err)
		return SessionItem{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	rows = state.QueryRow(username)
	identify := -1
	err = rows.Scan(&identify)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Commit()
			stdio.LogInfo(username, "用户身份未知")
			return SessionItem{}, stdio.GetEmptyErrorMessage()
		} else {
			_ = tx.Rollback()
			stdio.LogWarn(username, "数据库SQL指令执行失败", err)
			return SessionItem{}, stdio.GetErrorMessage(-500, "请求处理出错")
		}
	}
	tx.Commit()
	if identify < 0 {
		stdio.LogWarn(username, "用户身份获取失败", nil)
		return SessionItem{}, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	if item.Session != "" {
		return SessionItem{
			Session:   item.Session,
			Identify:  identify,
			Exist:     true,
			Effective: item.Effective == 1,
			Expired:   item.Expired < time.Now().Unix(),
		}, stdio.GetEmptyErrorMessage()
	} else {
		return SessionItem{
			Exist: false,
		}, stdio.GetEmptyErrorMessage()
	}
}

func (sessionManagerImpl sessionManagerImpl) Update(username string, password string, session string, identify int) stdio.MessagedError {
	exist, errMessage := SessionManager.CheckUserExist(username, "user_token")
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
		state, err = tx.Prepare("insert into `user_token` (`u_id`, `u_password` ,`u_session`, `u_session_expired`, `u_token_effective`) values (?, ?, ?, ?, 1)")
	} else {
		state, err = tx.Prepare("update `user_token` set `u_session`=?, `u_session_expired`=?, `u_token_effective`=1, `u_password`=? where `u_id`=?")
	}
	if err != nil {
		stdio.LogWarn(username, "数据库准备SQL指令失败", err)
		return stdio.GetErrorMessage(-500, "请求处理出错")
	}
	if !exist {
		_, err = state.Exec(username, password, session, time.Now().Unix()+1800)
	} else {
		_, err = state.Exec(session, time.Now().Unix()+1800, password, username)
	}
	if err == nil {
		tx.Commit()
		if !exist {
			stdio.LogVerbose(username, "向数据库插入新 ASP.NET_SessionId 成功")
		} else {
			stdio.LogVerbose(username, "向数据库更新 ASP.NET_SessionId 成功")
		}
	} else {
		_ = tx.Rollback()
		stdio.LogWarn(username, "数据库SQL指令执行失败", err)
		return stdio.GetErrorMessage(-500, "请求处理出错")
	}

	exist, errMessage = SessionManager.CheckUserExist(username, "user_info")
	if errMessage.HasInfo {
		return errMessage
	}
	tx, err = unit.Maria.Begin()
	if err != nil {
		stdio.LogWarn(username, "数据库开始事务失败", err)
		return stdio.GetErrorMessage(-500, "请求处理出错")
	}
	if !exist {
		state, err = tx.Prepare("insert into `user_info` (`u_id`, `u_identify`, `u_info_expired`) values (?, ?, 0)")
	} else {
		state, err = tx.Prepare("update `user_info` set `u_identify`=? where `u_id`=?")
	}
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn(username, "数据库准备SQL指令失败", err)
		return stdio.GetErrorMessage(-500, "请求处理出错")
	}
	if !exist {
		_, err = state.Exec(username, identify)
	} else {
		_, err = state.Exec(identify, username)
	}
	if err == nil {
		tx.Commit()
		if !exist {
			stdio.LogVerbose(username, "向数据库插入新用户身份成功")
		} else {
			stdio.LogVerbose(username, "向数据库更新用户身份成功")
		}
		return stdio.GetEmptyErrorMessage()
	} else {
		_ = tx.Rollback()
		stdio.LogWarn(username, "数据库查询失败", err)
		return stdio.GetErrorMessage(-500, "请求处理出错")
	}
}

func (sessionManagerImpl sessionManagerImpl) GetUserPassword(username string, password string) (string, stdio.MessagedError) {
	pass := password
	if pass == "" {
		tx, err := unit.Maria.Begin()
		if err != nil {
			stdio.LogWarn(username, "数据库开始事务失败", err)
			return "", stdio.GetErrorMessage(-500, "请求处理出错")
		}
		var state *sql.Stmt
		state, err = tx.Prepare("select `u_password` from `user_token` where `u_id`=?")
		if err != nil {
			_ = tx.Rollback()
			stdio.LogWarn(username, "数据库准备SQL指令失败", err)
			return "", stdio.GetErrorMessage(-500, "请求处理出错")
		}
		rows := state.QueryRow(username)
		err = rows.Scan(&pass)
		if err != nil {
			if err == sql.ErrNoRows {
				tx.Commit()
				return "", stdio.GetEmptyErrorMessage()
			} else {
				_ = tx.Rollback()
				stdio.LogWarn(username, "数据库SQL指令执行失败", err)
				return "", stdio.GetErrorMessage(-500, "请求处理出错")
			}
		}
		tx.Commit()
		if pass == "" {
			return "", stdio.GetEmptyErrorMessage()
		}
	}
	return pass, stdio.GetEmptyErrorMessage()
}

func (sessionManagerImpl sessionManagerImpl) CheckUserExist(username string, table string) (bool, stdio.MessagedError) {
	tx, err := unit.Maria.Begin()
	if err != nil {
		stdio.LogWarn(username, "数据库开始事务失败", err)
		return false, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	var state *sql.Stmt
	state, err = tx.Prepare("select `u_id` from `" + table + "` where `u_id`=?")
	if err != nil {
		_ = tx.Rollback()
		stdio.LogWarn(username, "数据库准备SQL指令失败", err)
		return false, stdio.GetErrorMessage(-500, "请求处理出错")
	}
	rows := state.QueryRow(username)
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
