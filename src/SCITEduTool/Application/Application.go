package Application

import (
	"SCITEduTool/Application/manager"
	"SCITEduTool/Application/stdio"
	"SCITEduTool/Application/unit"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type application interface {
	SetupWithConfig()
}

type applicationImpl struct{}

var Application application = applicationImpl{}

func (applicationImpl applicationImpl) SetupWithConfig() {
	stdio.LogInfo("", "工科助手API启动中")
	configDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		stdio.LogAssert("", "运行目录获取失败", err)
		os.Exit(0)
	}
	configDir += "/config"
	setupConfigDir(configDir)
	setupServer(configDir)
	setupToken(configDir)
	setupPrivateKey(configDir)
	stdio.LogInfo("", "工科助手API配置读取完成，配置文件将在重启生效，祝您使用愉快~")
}

func setupConfigDir(configDir string) {
	_, err := os.Stat(configDir)
	if err == nil {
		return
	}
	if os.IsNotExist(err) {
		err = os.MkdirAll(configDir, 0644)
		if err == nil {
			stdio.LogInfo("", "配置目录创建成功")
			return
		}
		stdio.LogAssert("", "配置目录创建失败", err)
	} else {
		stdio.LogAssert("", "配置目录信息失败", err)
	}
	os.Exit(0)
}

func setupServer(configDir string) {
	path := configDir + "/server.json"
	_, err := os.Stat(path)
	var file *os.File
	var sqlConfigContent []byte
	if err == nil {
		file, err = os.OpenFile(path, os.O_RDONLY, 0644)
		if err != nil {
			stdio.LogAssert("", "SQL配置读取失败", err)
			goto exit
		}
		sqlConfigContent, err = ioutil.ReadAll(io.Reader(file))
		_ = file.Close()
		sqlConf := unit.ServerConfig{}
		err = json.Unmarshal(sqlConfigContent, &sqlConf)
		if err != nil {
			stdio.LogAssert("", "SQL配置解析失败", err)
			goto exit
		}
		stdio.LocalDebug.SetupDebugConfig(sqlConf.Debug)
		unit.InitSQL(sqlConf.Sql)
		return
	}
	if os.IsNotExist(err) {
		sqlConf := unit.ServerConfig{
			Debug: false,
			Sql: unit.SqlConfig{
				Username: "//输入您的数据库用户名",
				Password: "//输入您的数据库密码",
				IP:       "//请输入您的数据库IP",
				Port:     "//请输入您的数据库监听端口",
				DBName:   "//请输入您的数据库用于工科助手的数据簿名称",
			},
		}
		sqlConfigContent, err = json.Marshal(sqlConf)
		err = ioutil.WriteFile(path, sqlConfigContent, 0644)
		if err == nil {
			stdio.LogAssert("", "SQL配置文件不存在，已为您新建默认配置文件，请修改后重新启动", nil)
		} else {
			stdio.LogAssert("", "默认SQL配置文件创建失败", err)
		}
		goto exit
	} else {
		stdio.LogAssert("", "配置目录信息失败", err)
	}
	stdio.LogAssert("", "SQL配置获取失败", err)

exit:
	os.Exit(0)
}

func setupToken(configDir string) {
	path := configDir + "/token.json"
	_, err := os.Stat(path)
	var file *os.File
	var tokenConfigContent []byte
	if err == nil {
		file, err = os.OpenFile(path, os.O_RDONLY, 0644)
		if err != nil {
			stdio.LogAssert("", "Token配置读取失败", err)
			goto exit
		}
		tokenConfigContent, err = ioutil.ReadAll(io.Reader(file))
		_ = file.Close()
		tokenConf := manager.TokenConfig{}
		err = json.Unmarshal(tokenConfigContent, &tokenConf)
		if err != nil {
			stdio.LogAssert("", "Token配置解析失败", err)
			goto exit
		}
		manager.TokenUnit.InitKey(tokenConf)
		return
	}
	if os.IsNotExist(err) {
		tokenConf := manager.TokenConfig{
			TokenKey: "//请输入您设定的TokenKey，建议设置为16位随机字符串。" +
				"此配置修改必然会导致当前所有用户token失效，请谨慎修改。",
			TokenSecret: "//请输入您设定的TokenSecret，建议设置为32位随机字符串。" +
				"此配置修改必然会导致当前所有用户token失效，请谨慎修改。",
			AccessExpired: "//请输入access_token过期时间，单位秒，默认2592000（30天）。" +
				"此配置修改可能会导致部分用户token失效，请谨慎修改。",
			RefreshExpired: "//请输入refresh_token过期时间，单位秒，默认124416000（4年）。" +
				"此配置修改可能会导致部分用户token失效，请谨慎修改。",
		}
		tokenConfigContent, err = json.Marshal(tokenConf)
		err = ioutil.WriteFile(path, tokenConfigContent, 0644)
		if err == nil {
			stdio.LogAssert("", "Token配置文件不存在，已为您新建默认配置文件，请修改后重新启动", nil)
		} else {
			stdio.LogAssert("", "默认Token配置文件创建失败", err)
		}
		goto exit
	} else {
		stdio.LogAssert("", "配置目录信息失败", err)
	}
	stdio.LogAssert("", "Token配置获取失败", err)

exit:
	os.Exit(0)
}

func setupPrivateKey(configDir string) {
	path := configDir + "/key.json"
	_, err := os.Stat(path)
	var file *os.File
	var keyConfigContent []byte
	if err == nil {
		file, err = os.OpenFile(path, os.O_RDONLY, 0644)
		if err != nil {
			stdio.LogAssert("", "Token配置读取失败", err)
			goto exit
		}
		keyConfigContent, err = ioutil.ReadAll(io.Reader(file))
		_ = file.Close()
		keyConf := unit.PrivateKey{}
		err = json.Unmarshal(keyConfigContent, &keyConf)
		if err != nil {
			stdio.LogAssert("", "Token配置解析失败", err)
			goto exit
		}
		unit.RSAStaticUnit.SetPrivateKey(keyConf)
		return
	}
	if os.IsNotExist(err) {
		tokenConf := unit.PrivateKey{
			Content: "//请粘贴您经过base64编码后的RSA私钥，仅支持RSA/ECB/PKCS1Padding，请勿直接将私钥粘贴至此。" +
				"此配置修改会导致所有用户登录状态失效，请谨慎修改，若需修改请提前清空数据表“user_token”。",
		}
		keyConfigContent, err = json.Marshal(tokenConf)
		err = ioutil.WriteFile(path, keyConfigContent, 0644)
		if err == nil {
			stdio.LogAssert("", "Token配置文件不存在，已为您新建默认配置文件，请修改后重新启动", nil)
		} else {
			stdio.LogAssert("", "默认Token配置文件创建失败", err)
		}
		goto exit
	} else {
		stdio.LogAssert("", "配置目录信息失败", err)
	}
	stdio.LogAssert("", "Token配置获取失败", err)

exit:
	os.Exit(0)
}
