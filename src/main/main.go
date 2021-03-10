package main

import (
	"SCITEduTool/base/Application"
)

func main() {
	Application.SetupWithConfig()
	Application.RegisterAPI()
}
