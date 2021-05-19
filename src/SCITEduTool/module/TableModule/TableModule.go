package TableModule

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"SCITEduTool/manager/ChartManager"
	"SCITEduTool/manager/InfoManager"
	"SCITEduTool/manager/TableManager"
	"SCITEduTool/module/InfoModule"
	"SCITEduTool/module/SessionModule"
	"SCITEduTool/unit/StdOutUnit"
	"github.com/PuerkitoBio/goquery"
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
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	resp.Body.Close()
	if err != nil {
		StdOutUnit.Error("", "HTML解析失败", err)
		return TableManager.TableObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	viewState := doc.Find("#__VIEWSTATE").AttrOr("value", "")
	if viewState == "" {
		StdOutUnit.Error(username, "未发现 __VIEWSTATE", nil)
		return TableManager.TableObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

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

	doc, err = goquery.NewDocumentFromReader(resp.Body)
	resp.Body.Close()
	if err != nil {
		StdOutUnit.Error("", "HTML解析失败", err)
		return TableManager.TableObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	className, errMessage := ChartManager.GetClassName(info.Faculty, info.Specialty, info.Class)
	if errMessage.HasInfo {
		return TableManager.TableObject{}, errMessage
	}
	tableId := ""
	selected := false
	doc.Find("#kb").Find("option").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if s.Text() == className {
			tableId = s.AttrOr("value", "")
			if s.AttrOr("selected", "nil") == "selected" {
				selected = true
			}
			return false
		} else {
			return true
		}
	})
	if tableId == "" {
		StdOutUnit.Error("", "tableId获取失败", err)
		return TableManager.TableObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	if selected {
		goto parse
	}
	viewState = doc.Find("#__VIEWSTATE").AttrOr("value", "")
	if viewState == "" {
		StdOutUnit.Error(username, "未发现 __VIEWSTATE", nil)
		return TableManager.TableObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

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

	doc, err = goquery.NewDocumentFromReader(resp.Body)
	resp.Body.Close()
	if err != nil {
		StdOutUnit.Error("", "HTML解析失败", err)
		return TableManager.TableObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	viewState = doc.Find("#__VIEWSTATE").AttrOr("value", "")
	if viewState == "" {
		StdOutUnit.Error(username, "未发现 __VIEWSTATE", nil)
		return TableManager.TableObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	selected = false
	doc.Find("#kb").Find("option").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if s.Text() == className {
			if s.AttrOr("selected", "nil") == "selected" {
				selected = true
			}
			return false
		} else {
			return true
		}
	})
	if !selected {
		StdOutUnit.Error("", "无法选中目标课表数据", err)
		return TableManager.TableObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

parse:
	tableObject := TableManager.TableObject{}
	resultCount := 0
	//dayFault := []int{0, 1, 0, 1, 0}
	hasError := false
	doc.Find("#Table6").Find("tbody").Find("tr").EachWithBreak(func(trIndex int, tr *goquery.Selection) bool {
		tr.Find("td").EachWithBreak(func(tdIndex int, td *goquery.Selection) bool {
			lesson := TableManager.LessonItem{
				Data: []TableManager.LessonSingleItem{},
			}
			html, err := td.Html()
			html = strings.ReplaceAll(html, "\n", "")
			if err != nil || html == " " {
				return !hasError
			}
			if td.AttrOr("rowspan", "0") != "2" ||
				len(strings.Split(html, "<br/>")) <= 1 {
				return !hasError
			}
			dayClassData := strings.Split(html, "<br/><br/><br/>")
			for _, data := range dayClassData {
				singleData := strings.Split(data, "<br/>")
				dataItem := TableManager.LessonSingleItem{}
				dataItem.Name = singleData[0]
				stringClass := singleData[1]
				stringClass = stringClass[:strings.Index(stringClass, "(")]
				stringClass = strings.ReplaceAll(stringClass, "单", "")
				stringClass = strings.ReplaceAll(stringClass, "双", "")
				StdOutUnit.Debug("", strconv.Itoa(tdIndex-1-trIndex/2%2)+", "+strconv.Itoa(trIndex/2-1)+": "+stringClass, nil)
				var rangeArray []string
				if strings.Contains(stringClass, ",") {
					rangeArray = strings.Split(stringClass, ",")
				} else {
					rangeArray = []string{stringClass}
				}
				weekRange0 := strings.Contains(singleData[1], "双")
				weekRange1 := strings.Contains(singleData[1], "单")

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
						hasError = true
						return !hasError
					}
					end, err := strconv.Atoi(localRange[1])
					if err != nil {
						StdOutUnit.Error(username, "课表解析失败", err)
						return !hasError
					}
					for index := start; index <= end; index++ {
						if (!weekRange0 && index/2*2 != index) || (!weekRange1 && index/2*2 == index) {
							dataItem.Range = append(dataItem.Range, index)
						}
					}
				}
				dataItem.Teacher = strings.ReplaceAll(singleData[2], "\n", "")
				dataItem.Room = singleData[3]
				lesson.Data = append(lesson.Data, dataItem)
			}
			resultCount++
			if (tdIndex-1-trIndex/2%2) >= 7 || (trIndex/2-1) >= 5 {
				hasError = true
				return !hasError
			}
			tableObject.Object[tdIndex-1-trIndex/2%2][trIndex/2-1] = lesson
			return true
		})
		return !hasError
	})

	if hasError {
		return TableManager.TableObject{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	if resultCount == 0 {
		StdOutUnit.Error(username, "课表数据为空", nil)
		return TableManager.TableObject{}, StdOutUnit.GetErrorMessage(-500, "课表数据为空")
	}

	for dayIndex, day := range tableObject.Object {
		for classIndex, class := range day {
			if class.Data != nil {
				continue
			}
			tableObject.Object[dayIndex][classIndex] = TableManager.LessonItem{
				Data: []TableManager.LessonSingleItem{},
			}
		}
	}
	TableManager.Update(username, info, year, semester, tableId, tableObject)
	return tableObject, StdOutUnit.GetEmptyErrorMessage()
}

func teacherTable(username string, info InfoManager.UserInfo, year string, semester int, session string) (TableManager.TableObject, StdOutUnit.MessagedError) {
	return TableManager.TableObject{}, StdOutUnit.GetErrorMessage(-500, "TODO")
}
