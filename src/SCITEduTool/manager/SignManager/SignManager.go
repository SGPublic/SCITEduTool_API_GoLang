package SignManager

import (
	"database/sql"

	"SCITEduTool/unit/SQLStaticUnit"
	"SCITEduTool/unit/StdOutUnit"
)

func GetAppSecretByAppKey(appKey string, platform string) string {
	tx, err := SQLStaticUnit.Maria.Begin()
	if err != nil {
		StdOutUnit.Warn("", "数据库开始事务失败", err)
		return ""
	}
	state, err := tx.Prepare("select `app_secret` from `sign_keys` where `app_key`=? and `platform`=?")
	if err != nil {
		_ = tx.Rollback()
		StdOutUnit.Warn("", "数据库准备SQL指令失败", err)
		return ""
	}
	rows := state.QueryRow(appKey, platform)
	secret := ""
	err = rows.Scan(&secret)
	if err == nil {
		tx.Commit()
		return secret
	}
	if err == sql.ErrNoRows {
		tx.Commit()
		return ""
	} else {
		_ = tx.Rollback()
		StdOutUnit.Warn("", "数据库SQL指令执行失败", err)
		return ""
	}
}

func GetDefaultAppSecretByPlatform(platform string) string {
	tx, err := SQLStaticUnit.Maria.Begin()
	if err != nil {
		StdOutUnit.Warn("", "数据库开始事务失败", err)
		return ""
	}
	state, err := tx.Prepare("select `app_secret` from `sign_keys` where `platform`=? order by `build` limit 1")
	if err != nil {
		_ = tx.Rollback()
		StdOutUnit.Warn("", "数据库准备SQL指令失败", err)
		return ""
	}
	rows := state.QueryRow(platform)
	secret := ""
	err = rows.Scan(&secret)
	if err == nil {
		tx.Commit()
		return secret
	}
	if err == sql.ErrNoRows {
		tx.Commit()
		return ""
	} else {
		_ = tx.Rollback()
		StdOutUnit.Warn("", "数据库SQL指令执行失败", err)
		return ""
	}
}
