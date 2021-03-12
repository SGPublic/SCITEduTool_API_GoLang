package AchieveManager

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"

	"SCITEduTool/manager/InfoManager"
	"SCITEduTool/manager/SessionManager"
	"SCITEduTool/unit/SQLStaticUnit"
	"SCITEduTool/unit/StdOutUnit"
)

type CurrentAchieveItem struct {
	Name       string `json:"name"`
	PaperScore string `json:"paper_score"`
	Mark       string `json:"mark"`
	Retake     string `json:"retake"`
	Rebuild    string `json:"rebuild"`
	Credit     string `json:"credit"`
}

type FailedAchieveItem struct {
	Name string `json:"name"`
	Mark string `json:"mark"`
}

type AchieveObject struct {
	Current []CurrentAchieveItem `json:"current"`
	Failed  []FailedAchieveItem  `json:"failed"`
}

type AchieveContent struct {
	Exist   bool
	Expired bool
	Achieve string
}

func Get(username string, grade int, schoolYear string, semester int) (AchieveContent, StdOutUnit.MessagedError) {
	tx, err := SQLStaticUnit.Maria.Begin()
	if err != nil {
		StdOutUnit.Warn(username, "数据库开始事务失败", err)
		return AchieveContent{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	achieveId := GetAchieveId(grade, schoolYear, semester)
	if achieveId == "" {
		StdOutUnit.Warn("", "成绩单编号解析失败", err)
		return AchieveContent{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	state, err := tx.Prepare("select `a_content_" + achieveId + "` from `student_achieve` where `u_id`=?")
	if err != nil {
		_ = tx.Rollback()
		StdOutUnit.Warn(username, "数据库准备SQL指令失败", err)
		return AchieveContent{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	rows := state.QueryRow(username)
	achieve := AchieveContent{}
	achieveString := sql.NullString{}
	err = rows.Scan(&achieveString)
	if err == nil {
		tx.Commit()
		achieve.Exist = achieveString.Valid
		if achieveString.Valid {
			achieve.Achieve = achieveString.String
		}
		return achieve, StdOutUnit.GetEmptyErrorMessage()
	}
	if err == sql.ErrNoRows {
		tx.Commit()
		return AchieveContent{}, StdOutUnit.GetEmptyErrorMessage()
	} else {
		_ = tx.Rollback()
		StdOutUnit.Warn(username, "数据库SQL指令执行失败", err)
		return AchieveContent{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
}

func Update(username string, info InfoManager.UserInfo, year string, semester int,
	achieve AchieveObject) StdOutUnit.MessagedError {
	achieveContent, _ := json.Marshal(achieve)
	achieveString := string(achieveContent)

	exist, errMessage := SessionManager.CheckUserExist(username, "student_achieve")
	if errMessage.HasInfo {
		return errMessage
	}
	tx, err := SQLStaticUnit.Maria.Begin()
	if err != nil {
		StdOutUnit.Warn(username, "数据库开始事务失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	achieveId := GetAchieveId(info.Grade, year, semester)
	if achieveId == "" {
		StdOutUnit.Warn("", "成绩单编号解析失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	var state *sql.Stmt
	if !exist {
		state, err = tx.Prepare("insert into `student_achieve` (`u_id`, `u_faculty` ,`u_specialty`, `u_class`, `u_grade`, `a_content_" +
			achieveId + "`) values (?, ?, ?, ?, ?, ?)")
	} else {
		state, err = tx.Prepare("update `student_achieve` set `a_content_" +
			achieveId + "`=? where `u_id`=?")
	}
	if err != nil {
		_ = tx.Rollback()
		StdOutUnit.Warn(username, "数据库准备SQL指令失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	if !exist {
		_, err = state.Exec(username, info.Faculty, info.Specialty, info.Class, info.Grade, achieveString)
	} else {
		_, err = state.Exec(achieveString, username)
	}
	if err != nil {
		_ = tx.Rollback()
		StdOutUnit.Warn(username, "数据库SQL指令执行失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	tx.Commit()
	if !exist {
		StdOutUnit.Verbose(username, "向数据库插入新成绩数据成功")
	} else {
		StdOutUnit.Verbose(username, "向数据库更新成绩数据成功")
	}
	return StdOutUnit.GetEmptyErrorMessage()
}

func GetAchieveId(grade int, schoolYear string, semester int) string {
	year := strings.Split(schoolYear, "-")
	yearStart, err := strconv.Atoi(year[0])
	if err != nil {
		return ""
	}
	result := "0" + strconv.Itoa((yearStart-grade)*2+semester)
	return result[len(result)-2:]
}
