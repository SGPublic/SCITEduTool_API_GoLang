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

var maria *sql.DB

func InitSQL() {
	var err error
	maria, err = sql.Open("mysql", strings.Join([]string{
		userName, ":", password,
		"@tcp(", ip, ":", port, ")/",
		dbName, "?charset=utf8",
	}, ""))
	if err != nil {
		StdOutUnit.Assert.String("", err.Error())
		os.Exit(0)
	}
	err = maria.Ping()
	if err != nil {
		StdOutUnit.Assert.String("", err.Error())
		os.Exit(0)
	}
}

func NewTransaction() (*sql.Tx, error) {
	if maria == nil {
		StdOutUnit.Assert.String("", "数据库未链接")
		os.Exit(0)
	}
	return maria.Begin()
}
