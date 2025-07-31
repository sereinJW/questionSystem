package config

import (
	"fmt"
	"os"
	"system_api/model"

	"github.com/joho/godotenv"
)

// 读取env文件
func Envinit() (model.Ai, error) {
	if err := godotenv.Load(); err != nil {
		return model.Ai{}, fmt.Errorf("读取env文件失败：%w", err)
	}

	var ai model.Ai
	ai = model.Ai{
		Api_key: os.Getenv("DEEPSEEK_API_KEY"),
		Url:     os.Getenv("DEEPSEEK_URL"),
		Model:   os.Getenv("DEEPSEEK_MODEL"),
	}

	if ai.Api_key == "" || ai.Url == "" || ai.Model == "" {
		return model.Ai{}, fmt.Errorf("env配置不完整")
	}

	return ai, nil
}
