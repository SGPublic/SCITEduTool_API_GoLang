package SQLStaticUnit

import (
	"database/sql"
	"os"
	"strconv"
	"strings"

	"SCITEduTool/unit/StdOutUnit"
	_ "github.com/go-sql-driver/mysql"
)

var Maria *sql.DB

type SQLConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	IP       string `json:"ip"`
	Port     string `json:"port"`
	DBName   string `json:"db_name"`
}

func InitSQL(conf SQLConfig) {
	var err error
	if strings.Contains(conf.Username, "//") ||
		conf.Username == "" || conf.Password == "" {
		StdOutUnit.Assert("", "用户名或密码为空或格式不正确", nil)
		os.Exit(0)
	}
	if strings.Contains(conf.IP, "//") || conf.IP == "" {
		StdOutUnit.Warn("", "数据库IP为空或格式不正确，将使用默认值", nil)
		conf.IP = "localhost"
	}
	_, err = strconv.Atoi(conf.Port)
	if err != nil {
		StdOutUnit.Warn("", "数据库端口为空或格式不正确，将使用默认值", nil)
		conf.Port = "3306"
	}
	if strings.Contains(conf.DBName, "//") || conf.DBName == "" {
		StdOutUnit.Warn("", "数据簿名称为空或格式不正确，将使用默认值", nil)
		conf.DBName = "scit_edu_tool"
	}
	Maria, err = sql.Open("mysql", strings.Join([]string{
		conf.Username, ":", conf.Password,
		"@tcp(", conf.IP, ":", conf.Port, ")/",
		conf.DBName, "?charset=utf8",
	}, ""))
	if err != nil {
		StdOutUnit.Assert("", "数据库模块初始化失败", err)
		os.Exit(0)
	}
	err = Maria.Ping()
	if err != nil {
		StdOutUnit.Assert("", "数据库连接失败", err)
		os.Exit(0)
	}
	StdOutUnit.Verbose("", "SQL配置成功")
}
