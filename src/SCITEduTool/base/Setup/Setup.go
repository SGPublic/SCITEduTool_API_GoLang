package Setup

import (
	"SCITEduTool/api"
	"SCITEduTool/unit/SQLStaticUnit"
	"log"
	"net/http"
)

func Do() {
	SQLStaticUnit.InitSQL()
	registerApi("/day", api.Day)
	registerApi("/getKey", api.GetKey)
	registerApi("/login", api.Login)
	registerApi("/info", api.Info)
	startService(":8000")
}

func registerApi(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	basePattern := "/api"
	http.HandleFunc(basePattern+pattern, handler)
}

func startService(addr string) {
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
