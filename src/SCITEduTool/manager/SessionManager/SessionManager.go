package SessionManager

import (
	"SCITEduTool/unit/SQLStaticUnit"
	"SCITEduTool/unit/StdOutUnit"
	"time"
)

type userSessionItem struct {
	Session string
	Expired int64
	Effective int
}

type SessionItem struct {
	Session string
	Exist bool
	Expired bool
	Effective bool
}

func Get(username string) (SessionItem, StdOutUnit.MessagedError) {
	tx, err := SQLStaticUnit.NewTransaction()
	if err != nil {
		StdOutUnit.Error.String(username, err.Error())
		return SessionItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	rows, err := tx.Query("select `u_session`,`u_session_expired`,`u_token_effective` from `user_info` where `u_id`=?",
		username)
	if err != nil {
		StdOutUnit.Error.String(username, err.Error())
		return SessionItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	var items []userSessionItem
	for rows.Next() {
		var item userSessionItem
		//err := rows.Scan(&item.Session, &item.Expired, &item.Effective)
		err := rows.Scan(&item)
		if err != nil {
			StdOutUnit.Error.String(username, err.Error())
			return SessionItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
		}
		items = append(items, item)
	}
	tx.Commit()
	if len(items) > 0 {
		item := items[0]
		return SessionItem{
			Session: item.Session,
			Exist: true,
			Effective: item.Effective == 1,
			Expired: item.Expired < time.Now().Unix(),
		}, StdOutUnit.GetEmptyErrorMessage()
	} else {
		return SessionItem {
			Exist: false,
		}, StdOutUnit.GetEmptyErrorMessage()
	}
}

func Update(username string, password string, session string) (bool, StdOutUnit.MessagedError) {
	tx, err := SQLStaticUnit.NewTransaction()
	if err != nil {
		StdOutUnit.Error.String(username, err.Error())
		return false, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	rows, err := tx.Query("select `u_id` from `user_info` where `u_id`=`?`",
		username)
	if err != nil {
		StdOutUnit.Error.String(username, err.Error())
		return false, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	var ids []string
	for rows.Next() {
		var id string
		//err := rows.Scan(&item.Session, &item.Expired, &item.Effective)
		err := rows.Scan(&id)
		if err != nil {
			StdOutUnit.Error.String(username, err.Error())
			return false, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
		}
		ids = append(ids, id)
	}
	tx.Commit()
	if len(ids) == 0 {
		tx, err := SQLStaticUnit.NewTransaction()
		if err != nil {
			StdOutUnit.Error.String(username, err.Error())
			return false, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
		}
		_, err = tx.Exec("insert into `user_info` (`u_id`, `u_password` ,`u_session`, `u_session_expired`, `u_token_effective`) values (`?`, `?`, `?`, `?`, `1`)",
			username, password, session, time.Now().Unix() + 1800)
		tx.Commit()
		if err != nil {
			StdOutUnit.Verbose.String(username, "向数据库插入新 ASP.NET_SessionId 成功")
			return true, StdOutUnit.GetEmptyErrorMessage()
		} else {
			StdOutUnit.Error.String(username, err.Error())
			return false, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
		}
	} else {
		_, err := tx.Exec("update `user_info` set `u_session`=`?`, `u_session_expired`=`?`, `u_token_effective`=`1`, `u_password`=?",
			session, time.Now().Unix() + 1800, password)
		tx.Commit()
		if err != nil {
			StdOutUnit.Verbose.String(username, "向数据库更新 ASP.NET_SessionId 成功")
			return true, StdOutUnit.GetEmptyErrorMessage()
		} else {
			return false, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
		}
	}
}
