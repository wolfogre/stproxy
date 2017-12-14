package main

import (
	"os"
	"os/exec"
	"log"
	"github.com/gin-gonic/gin"
	"net/http"
)

var (
	status = ""
)

func main() {
	log.SetFlags(log.LstdFlags | log.LstdFlags)
	if _, err := os.Stat("/opt/st"); os.IsNotExist(err) {
		err := exec.Command("git", "clone", "https://github.com/wolfogre/st.git", "/opt/st").Run()
		log.Panic(err)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(gin.LoggerWithWriter(gin.DefaultWriter, "/_status"))

	engine.GET("/_status", handleStatus)
	engine.POST("/_webhook", handleWebhook)

	engine.GET("/", handleIndex)
	engine.GET("/func/:name", handleFunc)

	engine.Run(":80")
}

func handleStatus(c *gin.Context) {
	if status != "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": status,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"msg": "OK",
		})
	}
}

func handleWebhook(c *gin.Context) {
	if c.GetHeader("X-GitHub-Event") != "push" {
		c.JSON(http.StatusOK, gin.H{
			"msg": "OK",
		})
		return
	}
	cmd := exec.Command("git", "pull")
	cmd.Dir = "/opt/st/"
	err := cmd.Run()
	if err != nil {
		status = err.Error()
	}
	c.JSON(http.StatusOK, gin.H{
		"msg": "OK",
	})
}

func handleIndex(c *gin.Context) {
	c.File("/opt/st/index.sh")
}

func handleFunc(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.AbortWithStatus(http.StatusNotFound)
	}
	c.File("/opt/st/" + name + ".sh")
}