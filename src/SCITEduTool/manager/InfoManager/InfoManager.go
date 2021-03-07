package InfoManager

import (
	"SCITEduTool/unit/SQLStaticUnit"
	"SCITEduTool/unit/StdOutUnit"
)

func GetUserPassword(username string) (string, StdOutUnit.MessagedError) {
	tx, err := SQLStaticUnit.NewTransaction()
	if err != nil {
		return "", StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	rows, err := tx.Query("select `u_password` from `user_info` where `u_id`=?", username)
	if err != nil {
		return "", StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	var pass []string
	for rows.Next() {
		var passItem string
		err = rows.Scan(passItem)
		if err != nil {
			return "", StdOutUnit.GetErrorMessage(-500, "请求处理出错")
		}
		pass = append(pass, passItem)
	}
	tx.Commit()
	if len(pass) > 0 {
		return pass[0], StdOutUnit.GetEmptyErrorMessage()
	} else {
		return "", StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
}