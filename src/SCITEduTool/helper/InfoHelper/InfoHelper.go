package InfoHelper

import (
	"SCITEduTool/helper/SessionHelper"
	"SCITEduTool/manager/ChartManager"
	"SCITEduTool/manager/InfoManager"
	"SCITEduTool/unit/StdOutUnit"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func Get(username string) (InfoManager.UserInfo, StdOutUnit.MessagedError) {
	info, errMessage := InfoManager.Get(username)
	if errMessage.HasInfo {
		return InfoManager.UserInfo{}, errMessage
	}
	if !info.Exist {
		StdOutUnit.Info.String(username, "用户名称不存在")
		return InfoManager.UserInfo{}, StdOutUnit.GetEmptyErrorMessage()
	}
	if !info.Expired {
		return info, StdOutUnit.GetEmptyErrorMessage()
	}
	StdOutUnit.Info.String(username, "用户基本名称过期")
	session, identify, errMessage := SessionHelper.Get(username, "")
	if errMessage.HasInfo {
		return InfoManager.UserInfo{}, errMessage
	}
	info, errMessage = Refresh(username, session, identify)
	if errMessage.HasInfo {
		return InfoManager.UserInfo{}, errMessage
	} else {
		return info, StdOutUnit.GetEmptyErrorMessage()
	}
}

func Refresh(username string, session string, identify int) (InfoManager.UserInfo, StdOutUnit.MessagedError) {
	switch identify {
	case 0:
		return studentInfo(username, session)
	case 1:
		return teacherInfo(username, session)
	default:
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
}

func studentInfo(username string, session string) (InfoManager.UserInfo, StdOutUnit.MessagedError) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, _ := http.NewRequest("GET", "http://218.6.163.93:8081/xsgrxx.aspx?xh="+username, nil)
	req.AddCookie(&http.Cookie{Name: "ASP.NET_SessionId", Value: session})
	resp, err := client.Do(req)
	if err != nil {
		StdOutUnit.Error.String(username, err.Error())
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	r, _ := regexp.Compile("__VIEWSTATE\" value=\"(.*?)\"")
	viewState := r.FindString(string(body))
	if viewState == "" {
		StdOutUnit.Error.String(username, "未发现 __VIEWSTATE")
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	r, _ = regexp.Compile("lbl_dqszj\">(.*?)<")
	grade_pre := r.FindString(string(body))
	if grade_pre == "" {
		StdOutUnit.Error.String(username, "年级名称获取失败")
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	grade_pre = grade_pre[11 : len(grade_pre)-1]
	grade, err := strconv.Atoi(grade_pre)
	if err != nil {
		StdOutUnit.Error.String(username, err.Error())
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	r, _ = regexp.Compile("xm\">(.*?)<")
	name := r.FindString(string(body))
	if name == "" {
		StdOutUnit.Error.String(username, "姓名名称获取失败")
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	name = name[5 : len(name)-1]

	r, _ = regexp.Compile("lbl_xzb\">(.*?)<")
	lbl_xzb := r.FindString(string(body))
	if lbl_xzb == "" {
		StdOutUnit.Error.String(username, "班级名称获取失败")
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	lbl_xzb = lbl_xzb[9 : len(lbl_xzb)-1]
	r, _ = regexp.Compile("(\\d+)\\.?(\\d+)班")
	class_pre := r.FindString(string(body))
	if class_pre == "" {
		StdOutUnit.Error.String(username, "班级ID获取失败")
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	class, err := strconv.Atoi(strings.ReplaceAll(class_pre, "班", ""))
	if err != nil {
		StdOutUnit.Error.String(username, "班级ID解析失败")
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	r, _ = regexp.Compile("lbl_xy\">(.*?)<")
	lbl_xy := r.FindString(string(body))
	if lbl_xy == "" {
		StdOutUnit.Error.String(username, "学院名称获取失败")
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	lbl_xy = lbl_xy[7 : len(lbl_xy)-1]

	r, _ = regexp.Compile("lbl_zymc\">(.*?)<")
	lbl_zymc_pre := r.FindString(string(body))
	if lbl_zymc_pre == "" {
		StdOutUnit.Error.String(username, "专业名称获取失败")
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	lbl_zymc_pre = lbl_zymc_pre[7 : len(lbl_zymc_pre)-1]
	lbl_zymc := strings.ReplaceAll(lbl_zymc_pre, "（", "(")

	req, _ = http.NewRequest("GET", "http://218.6.163.93:8081/tjkbcx.aspx?xh="+username, nil)
	req.AddCookie(&http.Cookie{Name: "ASP.NET_SessionId", Value: session})
	resp, err = client.Do(req)
	if err != nil {
		StdOutUnit.Error.String(username, err.Error())
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	body, err = ioutil.ReadAll(resp.Body)

	lbl_zymc_id := -1
	lbl_xy_id := -1

	r, _ = regexp.Compile("完成评价工作后")
	reeult := r.MatchString(string(body))
	if reeult {
		item, _ := ChartManager.GetChartIDWithClassName(lbl_xzb)
		if !item.Exist {
			return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
		}
		lbl_xy_id = item.FacultyId
		lbl_zymc_id = item.SpecialtyId
	} else {
		r, _ = regexp.Compile("value=\"(.*?)\">" + lbl_xy)
		lbl_xy_id_pre := r.FindString(string(body))
		if lbl_xy_id_pre == "" {
			StdOutUnit.Error.String(username, "学院ID获取失败")
			return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
		}
		lbl_xy_id, err = strconv.Atoi(lbl_xy_id_pre[7 : len(lbl_xy_id_pre)-len(lbl_xy)-1])
		if err != nil {
			StdOutUnit.Debug.String(username, "学院ID解析失败")
		}

		r, _ = regexp.Compile("value=\"(.*?)\">" + lbl_zymc_pre)
		lbl_zymc_id_pre := r.FindString(string(body))
		if lbl_zymc_id_pre == "" {
			StdOutUnit.Error.String(username, "专业ID获取失败")
			return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
		}
		lbl_zymc_id, err = strconv.Atoi(lbl_zymc_id_pre[7 : len(lbl_zymc_id_pre)-len(lbl_zymc_pre)-1])
		if err != nil {
			StdOutUnit.Debug.String(username, "专业ID解析失败")
		}
	}
	if lbl_zymc_id < 0 || lbl_xy_id < 0 {
		return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	ChartManager.WriteFacultyName(lbl_xy_id, lbl_xy)
	ChartManager.WriteSpecialtyName(lbl_xy_id, lbl_zymc_id, lbl_zymc)
	ChartManager.WriteClassName(lbl_xy_id, lbl_zymc_id, class, lbl_zymc)
	InfoManager.Update(username, name, lbl_xy_id, lbl_zymc_id, class, grade)
	return InfoManager.UserInfo{
		Name:      name,
		Faculty:   lbl_xy_id,
		Specialty: lbl_zymc_id,
		Class:     class,
		Grade:     grade,
	}, StdOutUnit.GetEmptyErrorMessage()
}

func teacherInfo(username string, session string) (InfoManager.UserInfo, StdOutUnit.MessagedError) {
	return InfoManager.UserInfo{}, StdOutUnit.GetErrorMessage(-500, "TODO")
}
