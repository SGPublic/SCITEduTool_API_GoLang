package unit

import (
	"SCITEduTool/Application/stdio"
	"database/sql"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var Maria *sql.DB

type ServerConfig struct {
	Debug bool      `json:"debug"`
	Sql   SqlConfig `json:"sql"`
}

type SqlConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	IP       string `json:"ip"`
	Port     string `json:"port"`
	DBName   string `json:"db_name"`
}

func InitSQL(conf SqlConfig) {
	var err error
	if strings.Contains(conf.Username, "//") ||
		conf.Username == "" || conf.Password == "" {
		stdio.LogAssert("", "用户名或密码为空或格式不正确", nil)
		os.Exit(0)
	}
	if strings.Contains(conf.IP, "//") || conf.IP == "" {
		stdio.LogWarn("", "数据库IP为空或格式不正确，将使用默认值", nil)
		conf.IP = "localhost"
	}
	_, err = strconv.Atoi(conf.Port)
	if err != nil {
		stdio.LogWarn("", "数据库端口为空或格式不正确，将使用默认值", nil)
		conf.Port = "3306"
	}
	if strings.Contains(conf.DBName, "//") || conf.DBName == "" {
		stdio.LogWarn("", "数据簿名称为空或格式不正确，将使用默认值", nil)
		conf.DBName = "scit_edu_tool"
	}
	Maria, err = sql.Open("mysql", strings.Join([]string{
		conf.Username, ":", conf.Password,
		"@tcp(", conf.IP, ":", conf.Port, ")/",
		conf.DBName, "?charset=utf8",
	}, ""))
	if err != nil {
		stdio.LogAssert("", "数据库模块初始化失败", err)
		os.Exit(0)
	}
	err = Maria.Ping()
	if err != nil {
		stdio.LogAssert("", "数据库连接失败", err)
		os.Exit(0)
	}
	stdio.LogVerbose("", "SQL配置成功")
}
