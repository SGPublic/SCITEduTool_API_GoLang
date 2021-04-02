package AchieveManager

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"SCITEduTool/base/LocalDebug"
	"SCITEduTool/manager/ChartManager"
	"SCITEduTool/manager/InfoManager"
	"SCITEduTool/unit/SQLStaticUnit"
	"SCITEduTool/unit/StdOutUnit"

	"github.com/xuri/excelize"
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

type TableExtractInfo struct {
	Name      string
	Data      []byte
	ErrorInfo string
}

func Get(username string, year string, semester int) (TableExtractInfo, StdOutUnit.MessagedError) {
	tx, err := SQLStaticUnit.Maria.Begin()
	if err != nil {
		StdOutUnit.Warn(username, "数据库开始事务失败", err)
		return TableExtractInfo{
			ErrorInfo: "服务器内部错误",
		}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	state, err := tx.Prepare("select `u_faculty`,`u_specialty`,`u_class`,`u_name` from `user_info` where `u_id`=?")
	if err != nil {
		_ = tx.Rollback()
		StdOutUnit.Warn(username, "数据库准备SQL指令失败", err)
		return TableExtractInfo{
			ErrorInfo: "服务器内部错误",
		}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	rows := state.QueryRow(username)
	info := InfoManager.UserInfo{}
	err = rows.Scan(&info.Faculty, &info.Specialty, &info.Class, &info.Name)
	if err != nil {
		_ = tx.Rollback()
		StdOutUnit.Warn(username, "数据库执行指令失败", err)
		return TableExtractInfo{
			ErrorInfo: "用户信息不存在",
		}, StdOutUnit.GetErrorMessage(-500, "请求处理失败")
	}
	baseDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		_ = tx.Rollback()
		StdOutUnit.Warn("", "运行目录获取失败", err)
		return TableExtractInfo{
			ErrorInfo: "服务器内部错误",
		}, StdOutUnit.GetErrorMessage(-500, "请求处理失败")
	}
	tx.Commit()
	baseDir += "/achieve"
	tableDir := baseDir + "/user/" + year + "/" + strconv.Itoa(semester) + "/" + strconv.Itoa(info.Faculty) + "/" + strconv.Itoa(info.Specialty) +
		"/" + strconv.Itoa(info.Class) + "/"
	_, err = os.Stat(tableDir)
	if err == nil {
		goto readFile
	}
	if os.IsNotExist(err) {
		StdOutUnit.Warn("", "成绩单目录不存在", err)
	} else {
		StdOutUnit.Warn("", "成绩单目录信息失败", err)
	}
	return TableExtractInfo{
		ErrorInfo: "目标成绩单不存在",
	}, StdOutUnit.GetErrorMessage(-500, "请求处理失败")

readFile:
	_, err = os.Stat(tableDir + username + ".xlsx")
	if err == nil {
		goto returnResult
	}
	if os.IsNotExist(err) {
		StdOutUnit.Warn("", "成绩单目录信息失败", err)
	} else {
		StdOutUnit.Warn("", "成绩单目录信息失败", err)
	}
	return TableExtractInfo{
		ErrorInfo: "目标成绩单不存在",
	}, StdOutUnit.GetErrorMessage(-500, "请求处理失败")

returnResult:
	table, err := ioutil.ReadFile(tableDir + username + ".xlsx")
	if err == nil {
		return TableExtractInfo{
			Name: info.Name,
			Data: table,
		}, StdOutUnit.GetEmptyErrorMessage()
	} else {
		StdOutUnit.Error("", "成绩单读取失败", err)
		return TableExtractInfo{
			ErrorInfo: "服务器内部错误",
		}, StdOutUnit.GetErrorMessage(-500, "请求处理失败")
	}
}

func Update(username string, info InfoManager.UserInfo, year string, semester int,
	achieve AchieveObject) StdOutUnit.MessagedError {
	var sample *os.File
	var target *os.File
	var table *excelize.File
	var creatTimePreString string
	var creatTimePre []string
	var creatTime time.Time

	baseDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		StdOutUnit.Warn("", "运行目录获取失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理失败")
	}
	baseDir += "/achieve"
	tableDir := baseDir + "/user/" + year
	if year != "all" {
		tableDir += "/" + strconv.Itoa(semester)
	}
	tableDir += "/" + strconv.Itoa(info.Faculty) + "/" + strconv.Itoa(info.Specialty) +
		"/" + strconv.Itoa(info.Class) + "/"
	_, err = os.Stat(tableDir)

	if err == nil {
		goto checkExist
	}
	if os.IsNotExist(err) {
		err = os.MkdirAll(tableDir, 0644)
		if err == nil {
			StdOutUnit.Info("", "成绩单目录创建成功")
			goto startExtract
		}
		StdOutUnit.Warn("", "成绩单目录创建失败", err)
	} else {
		StdOutUnit.Warn("", "成绩单目录信息失败", err)
	}
	return StdOutUnit.GetErrorMessage(-500, "请求处理失败")

checkExist:
	_, err = os.Stat(tableDir + username + ".xlsx")
	if err == nil {
		goto checkExtractTime
	}
	if os.IsNotExist(err) {
		goto startExtract
	} else {
		StdOutUnit.Warn("", "成绩单目录信息失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理失败")
	}

checkExtractTime:
	table, err = excelize.OpenFile(tableDir + username + ".xlsx")
	if err != nil {
		StdOutUnit.Warn("", "成绩单文件读取失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理失败")
	}

	creatTimePreString, err = table.GetCellValue("achieve", "D4")
	_ = table.Save()
	if err != nil {
		StdOutUnit.Warn("", "成绩单文件解析失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理失败")
	}

	creatTimePre = strings.Split(creatTimePreString, "：")
	if len(creatTimePre) != 2 {
		StdOutUnit.Warn("", "成绩单文件解析失败", err)
		goto startExtract
	}
	creatTime, err = time.Parse("2006年01月02日 15:04", creatTimePre[1])
	if err != nil {
		StdOutUnit.Warn("", "成绩单文件解析失败", err)
		goto startExtract
	}
	if creatTime.Sub(time.Now()).Hours() < 2 {
		return StdOutUnit.GetEmptyErrorMessage()
	}

startExtract:
	tableDir += username + ".xlsx"
	_ = os.Remove(tableDir)
	if LocalDebug.IsDebug() {
		sample, err = os.Open(baseDir + "\\achieve_sample.xlsx")
	} else {
		sample, err = os.Open(baseDir + "/achieve_sample.xlsx")
	}
	if err != nil {
		StdOutUnit.Warn("", "成绩单样本文件获取失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理失败")
	}
	if LocalDebug.IsDebug() {
		tableDir = strings.ReplaceAll(tableDir, "/", "\\")
	}
	target, err = os.OpenFile(tableDir, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		StdOutUnit.Warn("", "成绩单创建失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理失败")
	}
	_, err = io.Copy(target, sample)
	sample.Close()
	target.Close()
	if err != nil {
		StdOutUnit.Warn("", "成绩单文件复制失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理失败")
	}
	table, err = excelize.OpenFile(tableDir)
	if err != nil {
		StdOutUnit.Warn("", "成绩单文件解析失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理失败")
	}
	faculty, errMessage := ChartManager.GetFacultyName(info.Faculty)
	if errMessage.HasInfo {
		faculty = strconv.Itoa(info.Faculty)
	}
	specialty, errMessage := ChartManager.GetSpecialtyName(info.Faculty, info.Specialty)
	if errMessage.HasInfo {
		specialty = strconv.Itoa(info.Specialty)
	}
	class, errMessage := ChartManager.GetClassName(info.Faculty, info.Specialty, info.Class)
	if errMessage.HasInfo {
		class = strconv.Itoa(info.Class)
	}
	_ = table.SetCellStr("achieve", "A2", "姓名："+info.Name)
	_ = table.SetCellStr("achieve", "C2", "学号："+username)
	_ = table.SetCellStr("achieve", "A3", "学院："+faculty)
	_ = table.SetCellStr("achieve", "C3", "专业："+specialty)
	_ = table.SetCellStr("achieve", "E3", "行政班："+class)
	_ = table.SetCellStr("achieve", "A4", "学年："+year)
	if semester == 0 {
		_ = table.SetCellStr("achieve", "B4", "学期：（全年）")
	} else {
		_ = table.SetCellStr("achieve", "B4", "学期："+strconv.Itoa(semester))
	}
	_ = table.SetCellStr("achieve", "D4", "获取时间："+
		time.Now().Format("2006年01月02日 15:04"))
	for index, current := range achieve.Current {
		rowId := strconv.Itoa(6 + index)
		_ = table.SetCellStr("achieve", "A"+rowId, current.Name)
		_ = table.SetCellStr("achieve", "B"+rowId, current.PaperScore)
		_ = table.SetCellStr("achieve", "C"+rowId, current.Mark)
		_ = table.SetCellStr("achieve", "D"+rowId, current.Retake)
		_ = table.SetCellStr("achieve", "E"+rowId, current.Rebuild)
		_ = table.SetCellStr("achieve", "F"+rowId, current.Credit)
		_ = table.SetRowHeight("achieve", 6+index, 18.0)
		mark, err := strconv.Atoi(current.Mark)
		if err != nil {
			markFloat, err := strconv.ParseFloat(current.Mark, 32)
			if err != nil {
				mark = 0
			} else {
				mark = int(markFloat)
			}
		}
		retake, err := strconv.Atoi(current.Retake)
		if err != nil {
			markFloat, err := strconv.ParseFloat(current.Retake, 32)
			if err != nil {
				mark = 0
			} else {
				mark = int(markFloat)
			}
		}
		rebuild, err := strconv.Atoi(current.Rebuild)
		if err != nil {
			markFloat, err := strconv.ParseFloat(current.Rebuild, 32)
			if err != nil {
				mark = 0
			} else {
				mark = int(markFloat)
			}
		}
		pass := !(mark >= 60) && !(retake >= 60) && !(rebuild >= 60)
		if pass {
			_ = table.SetCellStyle("achieve", "A"+rowId, "A"+rowId, 12)
			_ = table.SetCellStyle("achieve", "B"+rowId, "F"+rowId, 13)
		} else {
			_ = table.SetCellStyle("achieve", "A"+rowId, "A"+rowId, 4)
			_ = table.SetCellStyle("achieve", "B"+rowId, "F"+rowId, 5)
		}
	}
	err = table.Save()
	if err != nil {
		StdOutUnit.Warn("", "成绩单文件保存失败", err)
		return StdOutUnit.GetErrorMessage(-500, "请求处理失败")
	} else {
		return StdOutUnit.GetEmptyErrorMessage()
	}
}
