package api

import (
	"net/http"
	"time"

	"SCITEduTool/unit/StdOutUnit"
)

const (
	Semester   = 2
	SchoolYear = "2020-2021"
	Evaluation = false
)

func Day(w http.ResponseWriter, r *http.Request) {
	base, errMessage := SetupAPI(w, r, nil)
	if errMessage.HasInfo {
		errMessage.OutMessage(w)
		return
	}
	timeStart := time.Date(2021, 2, 28, 0, 0, 0, 0, time.Local)
	timeNow := time.Now()

	left := timeNow.Sub(timeStart)

	StdOutUnit.Verbose("", "用户获取开学日期成功")
	base.OnObjectResult(struct {
		Code       int    `json:"code"`
		Message    string `json:"message"`
		DayCount   int    `json:"day_count"`
		Date       string `json:"date"`
		Semester   int    `json:"semester"`
		SchoolYear string `json:"school_year"`
		Evaluation bool   `json:"evaluation"`
	}{
		Code:       200,
		Message:    "success.",
		DayCount:   int(left.Hours() / 24),
		Date:       timeStart.In(time.Local).Format("2006/01/02"),
		Semester:   Semester,
		SchoolYear: SchoolYear,
		Evaluation: Evaluation,
	})
}
