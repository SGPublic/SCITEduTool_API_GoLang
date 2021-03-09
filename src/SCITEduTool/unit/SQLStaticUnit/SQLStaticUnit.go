package SQLStaticUnit

import (
	"database/sql"
	"os"
	"strings"

	"SCITEduTool/unit/StdOutUnit"
	_ "github.com/go-sql-driver/mysql"
)

const (
	userName = "root"
	password = "020821sky.."
	ip       = "localhost"
	port     = "3306"
	dbName   = "scit_edu_tool"
)

var Maria *sql.DB

func InitSQL() {
	var err error
	Maria, err = sql.Open("mysql", strings.Join([]string{
		userName, ":", password,
		"@tcp(", ip, ":", port, ")/",
		dbName, "?charset=utf8",
	}, ""))
	if err != nil {
		StdOutUnit.Assert.String("", err.Error())
		os.Exit(0)
	}
	err = Maria.Ping()
	if err != nil {
		StdOutUnit.Assert.String("", err.Error())
		os.Exit(0)
	}
}
