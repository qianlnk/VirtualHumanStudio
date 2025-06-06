package controllers

import (
	"github.com/qianlnk/VirtualHumanStudio/backend/client/promptt"
	"github.com/qianlnk/VirtualHumanStudio/backend/config"
)

var PrompttCli *promptt.Promptt

func InitPromptt() {
	PrompttCli = promptt.New(config.AppConfig.Promptt)
}
