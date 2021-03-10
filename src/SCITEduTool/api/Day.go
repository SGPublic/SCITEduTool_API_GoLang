package api

import (
	"SCITEduTool/unit/StdOutUnit"
	"net/http"
	"time"
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

	StdOutUnit.Verbose.String("", "用户获取开学日期成功")
	base.OnObjectResult(struct {
		Code       int
		Message    string
		DayCount   int
		Date       string
		Semester   int
		SchoolYear string
		Evaluation bool
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
