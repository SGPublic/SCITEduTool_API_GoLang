package TableModule

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"SCITEduTool/manager/InfoManager"
	"SCITEduTool/manager/TableManager"
	"SCITEduTool/module/InfoModule"
	"SCITEduTool/module/SessionModule"
	"SCITEduTool/unit/StdOutUnit"
)

func Get(username string, year string, semester int) (TableManager.TableObject, StdOutUnit.MessagedError) {
	info, errMessage := InfoModule.Get(username)
	if errMessage.HasInfo {
		return TableManager.TableObject{}, errMessage
	}
	table, errMessage := TableManager.Get(username, info, year, semester)
	if errMessage.HasInfo {
		return TableManager.TableObject{}, errMessage
	}
	if !table.Exist {
		goto refresh
	}
	if !table.Expired {
		var object = TableManager.TableObject{}
		err := json.Unmarshal([]byte(table.Table), &object)
		if err != nil {
			return TableManager.TableObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
		}
		return object, StdOutUnit.GetEmptyErrorMessage()
	}

refresh:
	session, _, errMessage := SessionModule.Get(username, "")
	if errMessage.HasInfo {
		return TableManager.TableObject{}, errMessage
	}
	tableContent, errMessage := Refresh(username, info, year, semester, session)
	if errMessage.HasInfo {
		return TableManager.TableObject{}, errMessage
	} else {
		return tableContent, StdOutUnit.GetEmptyErrorMessage()
	}
}

func Refresh(username string, info InfoManager.UserInfo, year string, semester int, session string) (TableManager.TableObject, StdOutUnit.MessagedError) {
	switch info.Identify {
	case 0:
		return studentTable(username, info, year, semester, session)
	case 1:
		return teacherTable(username, info, year, semester, session)
	default:
		return TableManager.TableObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
}

func studentTable(username string, info InfoManager.UserInfo, year string, semester int, session string) (TableManager.TableObject, StdOutUnit.MessagedError) {
	tableId := TableManager.GetTableId(info.Specialty, info.Grade, info.Class, year, semester)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	urlString := "http://218.6.163.93:8081/tjkbcx.aspx?xh=" + username
	req, _ := http.NewRequest("GET", urlString, nil)
	req.AddCookie(&http.Cookie{Name: "ASP.NET_SessionId", Value: session})
	req.Header.Add("Referer", urlString)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		StdOutUnit.Error(username, "网络请求失败", err)
		return TableManager.TableObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	r, _ := regexp.Compile("__VIEWSTATE\" value=\"(.*?)\"")
	viewState := r.FindString(string(body))
	if viewState == "" {
		StdOutUnit.Error(username, "未发现 __VIEWSTATE", nil)
		return TableManager.TableObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	viewState = viewState[20 : len(viewState)-1]

	form := url.Values{}
	form.Set("__EVENTTARGET", "zy")
	form.Set("__EVENTARGUMENT", "")
	form.Set("__LASTFOCUS", "")
	form.Set("__VIEWSTATE", viewState)
	form.Set("__VIEWSTATEGENERATOR", "3189F21D")
	form.Set("xn", year)
	form.Set("xq", strconv.Itoa(semester))
	form.Set("nj", strconv.Itoa(info.Grade))
	form.Set("xy", strconv.Itoa(info.Faculty))
	form.Set("zy", strconv.Itoa(info.Specialty))
	form.Set("kb", "")
	req, _ = http.NewRequest("POST", urlString, strings.NewReader(strings.TrimSpace(form.Encode())))
	req.AddCookie(&http.Cookie{Name: "ASP.NET_SessionId", Value: session})
	req.Header.Add("Referer", urlString)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err = client.Do(req)
	if err != nil {
		StdOutUnit.Error(username, "网络请求失败", err)
		return TableManager.TableObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	body, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	r, _ = regexp.Compile("selected=\"selected\" value=\"" + tableId + "\"")
	if r.MatchString(string(body)) {
		goto parse
	}
	r, _ = regexp.Compile("__VIEWSTATE\" value=\"(.*?)\"")
	viewState = r.FindString(string(body))
	if viewState == "" {
		StdOutUnit.Error(username, "未发现 __VIEWSTATE", nil)
		return TableManager.TableObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	viewState = viewState[20 : len(viewState)-1]

	form = url.Values{}
	form.Set("__EVENTTARGET", "kb")
	form.Set("__EVENTARGUMENT", "")
	form.Set("__LASTFOCUS", "")
	form.Set("__VIEWSTATE", viewState)
	form.Set("__VIEWSTATEGENERATOR", "3189F21D")
	form.Set("xn", year)
	form.Set("xq", strconv.Itoa(semester))
	form.Set("nj", strconv.Itoa(info.Grade))
	form.Set("xy", strconv.Itoa(info.Faculty))
	form.Set("zy", strconv.Itoa(info.Specialty))
	form.Set("kb", tableId)
	req, _ = http.NewRequest("POST", urlString, strings.NewReader(strings.TrimSpace(form.Encode())))
	req.AddCookie(&http.Cookie{Name: "ASP.NET_SessionId", Value: session})
	req.Header.Add("Referer", urlString)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err = client.Do(req)
	if err != nil {
		StdOutUnit.Error(username, "网络请求失败", err)
		return TableManager.TableObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	body, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	r, _ = regexp.Compile("__VIEWSTATE\" value=\"(.*?)\"")
	viewState = r.FindString(string(body))
	if viewState == "" {
		StdOutUnit.Error(username, "未发现 __VIEWSTATE", nil)
		return TableManager.TableObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	r, _ = regexp.Compile("selected=\"selected\" value=\"" + tableId + "\"")
	if !r.MatchString(string(body)) {
		StdOutUnit.Error(username, "无法选中目标课表数据", nil)
		return TableManager.TableObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

parse:
	bodyString := string(body)
	bodyString = strings.ReplaceAll(bodyString, "（", "(")
	bodyString = strings.ReplaceAll(bodyString, "）", ")")

	matchesClass := strings.Split(bodyString, "<tr>")

	tableObject := TableManager.TableObject{}
	resultCount := 0
	dayFault := []int{0, 1, 0, 1, 0}
	for class := 1; class < 6; class++ {
		matchesDay := strings.Split(matchesClass[class*2+1], "</td>")
		for day := 2; day < 8; day++ {
			lesson := TableManager.LessonItem{
				Data: []TableManager.LessonSingleItem{},
			}
			dayClassString := matchesDay[day-dayFault[class-1]]
			if dayClassString == "<td align=\"Center\">&nbsp;" || dayClassString == "<td align=\"Center\" width=\"7%\">&nbsp;" {
				tableObject.Object[day-2][class-1] = lesson
				continue
			}
			dayClassData := strings.Split(dayClassString, "<br>")
			dayClassDataCount := (len(dayClassData) + 1) / 7
			for count := 0; count < dayClassDataCount; count++ {
				dataItem := TableManager.LessonSingleItem{}
				classIndex := count * 7
				if count == 0 {
					dataItem.Name = strings.Split(dayClassData[classIndex], ">")[1]
				} else {
					dataItem.Name = dayClassData[classIndex]
				}
				stringClass := dayClassData[classIndex+1]
				stringClass = stringClass[0:strings.Index(stringClass, "(")]
				stringClass = strings.ReplaceAll(stringClass, "单", "")
				stringClass = strings.ReplaceAll(stringClass, "双", "")
				var rangeArray []string
				if strings.Contains(stringClass, ",") {
					rangeArray = strings.Split(stringClass, ",")
				} else {
					rangeArray = []string{stringClass}
				}
				weekRange0 := strings.Contains(dayClassData[classIndex+1], "双")
				weekRange1 := strings.Contains(dayClassData[classIndex+1], "单")

				for _, item := range rangeArray {
					var localRange []string
					if strings.Contains(item, "-") {
						localRange = strings.SplitN(item, "-", 2)
					} else {
						localRange = []string{item, item}
					}
					start, err := strconv.Atoi(localRange[0])
					if err != nil {
						StdOutUnit.Error(username, "课表解析失败", err)
						return TableManager.TableObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
					}
					end, err := strconv.Atoi(localRange[1])
					if err != nil {
						StdOutUnit.Error(username, "课表解析失败", err)
						return TableManager.TableObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
					}
					for index := start; index < end; index++ {
						if (!weekRange0 && index/2*2 != index) || (!weekRange1 && index/2*2 == index) {
							dataItem.Range = append(dataItem.Range, index)
						}
					}
				}
				dataItem.Teacher = strings.ReplaceAll(dayClassData[classIndex+2], "\n", "")
				dataItem.Room = dayClassData[classIndex+3]
				lesson.Data = append(lesson.Data, dataItem)
			}
			resultCount++
			tableObject.Object[day-2][class-1] = lesson
		}
	}
	if resultCount == 0 {
		StdOutUnit.Error(username, "课表数据为空", nil)
		return TableManager.TableObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	TableManager.Update(username, info, year, semester, tableObject)
	return tableObject, StdOutUnit.GetEmptyErrorMessage()
}

func teacherTable(username string, info InfoManager.UserInfo, year string, semester int, session string) (TableManager.TableObject, StdOutUnit.MessagedError) {
	return TableManager.TableObject{}, StdOutUnit.GetErrorMessage(-500, "TODO")
}
