package router

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"system_api/api"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func SetupRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()

	// 获取静态文件的绝对路径
	clientDistPath := "../client/dist"
	absPath, err := filepath.Abs(clientDistPath)
	if err != nil {
		fmt.Printf("获取静态文件路径失败: %v\n", err)
	} else {
		fmt.Printf("静态文件路径: %s\n", absPath)
	}

	// 检查路径是否存在
	if _, err := os.Stat(clientDistPath); os.IsNotExist(err) {
		fmt.Printf("警告: 静态文件目录不存在: %s\n", clientDistPath)
	} else {
		fmt.Printf("静态文件目录存在: %s\n", clientDistPath)
	}

	// 提供前端静态文件
	r.Static("/assets", filepath.Join(clientDistPath, "assets"))
	r.StaticFile("/", filepath.Join(clientDistPath, "index.html"))
	r.StaticFile("/index.html", filepath.Join(clientDistPath, "index.html"))
	r.StaticFile("/vite.svg", filepath.Join(clientDistPath, "vite.svg"))

	// 添加路由处理404请求
	r.NoRoute(func(c *gin.Context) {
		// 对于任何未匹配的路由都返回index.html
		c.File(filepath.Join(clientDistPath, "index.html"))
	})

	v, ok := binding.Validator.Engine().(*validator.Validate)
	if ok {
		validate := validator.New()
		*v = *validate
	}

	// API 路由
	apiRoutes := r.Group("/api")
	{
		questions := apiRoutes.Group("/questions")
		{
			questions.POST("/create", func(c *gin.Context) {
				api.CreateQuestion(c, db)
			})
			questions.GET("", func(c *gin.Context) {
				api.GetQuestions(c, db)
			})
			questions.POST("/add", func(c *gin.Context) {
				api.AddQuestion(c, db)
			})
			questions.POST("/edit", func(c *gin.Context) {
				api.EditQuestion(c, db)
			})
			questions.POST("/delete", func(c *gin.Context) {
				api.DeleteQuestion(c, db)
			})
		}
	}

	return r
}
