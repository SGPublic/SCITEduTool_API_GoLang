package NewsModule

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"SCITEduTool/manager/NewsManager"
	"SCITEduTool/unit/StdOutUnit"

	"github.com/PuerkitoBio/goquery"
)

func ListNewsByType(tid int, page int) ([]NewsManager.NewsItem, bool, StdOutUnit.MessagedError) {
	exist, errMessage := NewsManager.CheckChartExist(tid)
	if errMessage.HasInfo {
		StdOutUnit.Info("", "新闻类别查询失败")
		return nil, false, errMessage
	}
	if !exist {
		StdOutUnit.Info("", "新闻类别不存在")
		return nil, false, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	pageIndex := strconv.Itoa(page/2 + 1)
	urlString := "http://www.scit.cn/newslist" + strconv.Itoa(tid) + "_" + pageIndex + ".htm"
	req, _ := http.NewRequest("GET", urlString, nil)
	resp, err := client.Do(req)
	if err != nil {
		StdOutUnit.Error("", "网络请求失败", err)
		return nil, false, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	resp.Body.Close()
	if err != nil {
		StdOutUnit.Error("", "HTML解析失败", err)
		return nil, false, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	items := make([]NewsManager.NewsItem, 0)
	indexStart := 10 * (page % 2)
	hasNext := false
	r, _ := regexp.Compile("_(\\d*)\\.")
	doc.Find(".newslist").Find("ul").Find("li").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if i < indexStart {
			return true
		}
		if i == indexStart+11 {
			hasNext = true
			return false
		}
		idPre := r.FindString(s.Find("a").AttrOr("href", ""))
		if idPre == "" {
			StdOutUnit.Warn("", "tid获取失败，url: "+urlString+", index: "+strconv.Itoa(i), err)
			return true
		}
		if len(idPre) <= 2 {
			StdOutUnit.Warn("", "tid获取失败，url: "+urlString+", index: "+strconv.Itoa(i), err)
			return true
		}
		id, err := strconv.Atoi(idPre[1 : len(idPre)-1])
		if err != nil {
			StdOutUnit.Warn("", "tid获取失败，url: "+urlString+", index: "+strconv.Itoa(i), err)
			return true
		}
		item, errMessage := GetNewsById(tid, id)
		if errMessage.HasInfo {
			return true
		}

		items = append(items, item)
		return true
	})
	if doc.Find(".current").Text() != pageIndex {
		StdOutUnit.Debug("", "current: "+doc.Find(".current").Text(), nil)
		return make([]NewsManager.NewsItem, 0), false, StdOutUnit.GetEmptyErrorMessage()
	}
	doc.Find(".manu").EachWithBreak(func(_ int, s1 *goquery.Selection) bool {
		if s1.AttrOr("valign", "null") != "bottom" {
			return true
		}
		if s1.AttrOr("nowrap", "null") != "true" {
			return true
		}

		s1.Find("a").EachWithBreak(func(_ int, s2 *goquery.Selection) bool {
			if s2.Text() == "Next  > " {
				hasNext = s2.AttrOr("disabled", "null") != "disabled"
				return false
			}
			return true
		})
		return true
	})

	return items, hasNext, StdOutUnit.GetEmptyErrorMessage()
}

func GetTypeChart() ([]NewsManager.NewsTypeChartItem, StdOutUnit.MessagedError) {
	var items []NewsManager.NewsTypeChartItem
	var errMessage StdOutUnit.MessagedError
	items, errMessage = NewsManager.GetTypeChart()
	if errMessage.HasInfo {
		return nil, errMessage
	}
	if len(items) == 0 {
		items, errMessage = RefreshTypeChart()
		if errMessage.HasInfo {
			return nil, errMessage
		}
	}
	return items, StdOutUnit.GetEmptyErrorMessage()
}

func RefreshTypeChart() ([]NewsManager.NewsTypeChartItem, StdOutUnit.MessagedError) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	urlString := "http://m.scit.cn/news.aspx"
	req, _ := http.NewRequest("GET", urlString, nil)
	resp, err := client.Do(req)
	if err != nil {
		StdOutUnit.Error("", "网络请求失败", err)
		return nil, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	resp.Body.Close()
	if err != nil {
		StdOutUnit.Error("", "HTML解析失败", err)
		return nil, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	var charts []NewsManager.NewsTypeChartItem
	doc.Find(".menu").Find("ul").Find("li").Each(func(i int, s *goquery.Selection) {
		item := NewsManager.NewsTypeChartItem{
			Out: 1,
		}
		href := s.Find("a").AttrOr("href", "")
		r, _ := regexp.Compile("tid=(\\d+)")
		tidPre := r.FindString(href)
		if len(tidPre) <= 4 {
			StdOutUnit.Error("", "tid获取失败", err)
			return
		}
		tid, err := strconv.Atoi(tidPre[4:])
		if err != nil {
			StdOutUnit.Error("", "tid解析失败", err)
			return
		}
		item.TypeName = s.Find("a").Text()
		item.TypeId = tid
		charts = append(charts, item)
	})
	NewsManager.UpdateTypeChart(charts)
	return charts, StdOutUnit.GetEmptyErrorMessage()
}

func GetNewsById(tid int, id int) (NewsManager.NewsItem, StdOutUnit.MessagedError) {
	exist, errMessage := NewsManager.CheckNewsExist(tid, id)
	if errMessage.HasInfo {
		return NewsManager.NewsItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	if !exist {
		return RefreshNews(tid, id)
	}
	item, errMessage := NewsManager.GetNewsById(tid, id)
	if errMessage.HasInfo {
		return NewsManager.NewsItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	} else {
		return item, StdOutUnit.GetEmptyErrorMessage()
	}
}

func RefreshNews(tid int, id int) (NewsManager.NewsItem, StdOutUnit.MessagedError) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	urlString := "http://www.scit.cn/newsli" + strconv.Itoa(tid) + "_" + strconv.Itoa(id) + ".htm"
	req, _ := http.NewRequest("GET", urlString, nil)
	resp, err := client.Do(req)
	if err != nil {
		StdOutUnit.Error("", "网络请求失败", err)
		return NewsManager.NewsItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	resp.Body.Close()
	if err != nil {
		StdOutUnit.Error("", "HTML解析失败", err)
		return NewsManager.NewsItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	item := NewsManager.NewsItem{}
	item.Title = doc.Find(".news_title").Text()
	if item.Title == "" {
		StdOutUnit.Debug("", "新闻标题解析失败："+urlString, err)
		return NewsManager.NewsItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	item.Title = strings.ReplaceAll(item.Title, "\t", "")
	item.Title = strings.ReplaceAll(item.Title, "\n", "")
	r, _ := regexp.Compile("  ")
	for r.MatchString(item.Title) {
		item.Title = strings.ReplaceAll(item.Title, "  ", " ")
	}
	if item.Title[len(item.Title)-2:] == " " {
		item.Title = item.Title[:len(item.Title)-2]
	}
	item.Title = strings.ReplaceAll(item.Title, " ", " ")
	newsTimePre := doc.Find(".news_time")
	if newsTimePre.Text() == "" {
		StdOutUnit.Debug("", "新闻创建时间获取失败："+urlString, err)
		return NewsManager.NewsItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	newsTime := strings.Split(newsTimePre.Text(), " ")
	if len(newsTime) < 1 {
		StdOutUnit.Debug("", "新闻创建时间解析失败："+urlString, err)
		return NewsManager.NewsItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}
	item.CreateTime = newsTime[0]
	newsText := doc.Find(".news_text")
	item.Images = make([]string, 0)
	newsText.Find("p").Each(func(i int, s *goquery.Selection) {
		var img string
		img = s.Find("img").AttrOr("src", "")
		if img != "" {
			if len(item.Images) >= 3 {
				return
			}
			if img[:1] == "/" {
				img = "http://www.scit.cn" + img
			}
			item.Images = append(item.Images, img)
			return
		}

		if item.Summary != "" {
			return
		}
		if s.AttrOr("align", "") == "center" {
			return
		}
		summary := []rune(s.Text())
		if len(summary) > 80 {
			summary = summary[0:80]
		}
		item.Summary = strings.ReplaceAll(string(summary), "\t", "")
		item.Summary = strings.ReplaceAll(item.Summary, "\n", "")
		r, _ := regexp.Compile("  ")
		for r.MatchString(item.Summary) {
			item.Summary = strings.ReplaceAll(item.Summary, "  ", " ")
		}
		if len(item.Summary) < 2 {
			return
		}
		if item.Summary[len(item.Summary)-2:] == " " {
			item.Summary = item.Summary[:len(item.Summary)-2]
		}
		item.Summary = strings.ReplaceAll(item.Summary, " ", " ")
	})
	newsText.Find("img").Each(func(i int, s *goquery.Selection) {
		if len(item.Images) >= 3 {
			return
		}
		img := s.AttrOr("src", "")
		if img != "" {
			if img[:1] == "/" {
				img = "http://www.scit.cn" + img
			}
			exist := false
			for _, imageItem := range item.Images {
				if imageItem == img {
					exist = true
					break
				}
			}
			if !exist {
				item.Images = append(item.Images, img)
			}
		}
	})
	if item.Title == "" {
		StdOutUnit.Error("", "新闻标题获取失败，tid: "+strconv.Itoa(tid)+", nid: "+strconv.Itoa(id), err)
		return NewsManager.NewsItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	} else if item.Summary == "" && len(item.Images) == 0 {
		StdOutUnit.Error("", "新闻简介获取失败，tid: "+strconv.Itoa(tid)+", nid: "+strconv.Itoa(id), err)
		return NewsManager.NewsItem{}, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	} else {
		item.Tid = tid
		item.Nid = id
		NewsManager.UpdateNews(item)
		return item, StdOutUnit.GetEmptyErrorMessage()
	}
}

func GetHeadlines() ([]NewsManager.NewsItem, StdOutUnit.MessagedError) {
	headlines, errMessage := NewsManager.GetHeadlines()
	if errMessage.HasInfo {
		return nil, errMessage
	}
	var headline = headlines.News
	if headlines.Exist && !headlines.Expired {
		goto result
	}
	StdOutUnit.Debug("", "头条数据待更新", nil)
	headline, errMessage = RefreshHeadlines()
	if errMessage.HasInfo {
		return nil, errMessage
	}
result:
	news := make([]NewsManager.NewsItem, 0)
	for _, item := range headline {
		if item.Title == "" || item.Summary == "" {
			newsItem, errMessage := GetNewsById(item.Tid, item.Nid)
			if errMessage.HasInfo {
				return nil, errMessage
			}
			item.Title = newsItem.Title
			item.Summary = newsItem.Summary
		}
		news = append(news, item)
	}
	return news, StdOutUnit.GetEmptyErrorMessage()
}

func RefreshHeadlines() ([]NewsManager.NewsItem, StdOutUnit.MessagedError) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	urlString := "http://m.scit.cn/"
	req, _ := http.NewRequest("GET", urlString, nil)
	resp, err := client.Do(req)
	if err != nil {
		StdOutUnit.Error("", "网络请求失败", err)
		return nil, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	resp.Body.Close()
	if err != nil {
		StdOutUnit.Error("", "HTML解析失败", err)
		return nil, StdOutUnit.GetErrorMessage(-500, "请求处理出错")
	}

	var headlines []NewsManager.NewsItem
	doc.Find(".index_top_news_newslist").Each(func(i int, s *goquery.Selection) {
		item := NewsManager.NewsItem{}
		content := s.Find("a")
		href := content.AttrOr("href", "")
		r, _ := regexp.Compile("tid=(\\d+)")
		tidPre := r.FindString(href)
		if len(tidPre) <= 2 {
			StdOutUnit.Error("", "tid获取失败", err)
			return
		}
		tid, err := strconv.Atoi(tidPre[4:])
		if err != nil {
			StdOutUnit.Error("", "tid解析失败", err)
			return
		}
		item.Tid = tid
		r, _ = regexp.Compile("[^t]id=(\\d+)")
		idPre := r.FindString(href)
		if len(idPre) <= 4 {
			StdOutUnit.Error("", "id获取失败", err)
			return
		}
		id, err := strconv.Atoi(idPre[4:])
		if err != nil {
			StdOutUnit.Error("", "id解析失败", err)
			return
		}
		item.Nid = id
		item.Images = []string{
			content.Find(".index_top_news_newslist_img").Find("img").AttrOr("src", ""),
		}

		headlines = append(headlines, item)
	})

	doc.Find(".index_gkyw_news_item_one").Each(func(i int, s *goquery.Selection) {
		item := NewsManager.NewsItem{}
		img := s.Find(".index_gkyw_news_item_img").Find("a")
		href := img.AttrOr("href", "")
		r, _ := regexp.Compile("tid=(\\d+)")
		tidPre := r.FindString(href)
		if len(tidPre) <= 4 {
			StdOutUnit.Error("", "tid获取失败", err)
			return
		}
		tid, err := strconv.Atoi(tidPre[4:])
		if err != nil {
			StdOutUnit.Error("", "tid解析失败", err)
			return
		}
		item.Tid = tid
		r, _ = regexp.Compile("[^t]id=(\\d+)")
		idPre := r.FindString(href)
		if len(idPre) <= 4 {
			StdOutUnit.Error("", "id获取失败", err)
			return
		}
		id, err := strconv.Atoi(idPre[4:])
		if err != nil {
			StdOutUnit.Error("", "id解析失败", err)
			return
		}
		item.Nid = id
		item.Images = []string{
			img.Find("img").AttrOr("src", ""),
		}

		headlines = append(headlines, item)
	})
	NewsManager.UpdateHeadlines(headlines)
	return headlines, StdOutUnit.GetEmptyErrorMessage()
}
