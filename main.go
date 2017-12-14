package main

import (
	"os"
	"os/exec"
	"log"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	REPO_PATH = "/opt/st"
)

var (
	status = ""
)

func main() {
	log.SetFlags(log.LstdFlags | log.LstdFlags)
	if _, err := os.Stat(REPO_PATH); os.IsNotExist(err) {
		err := exec.Command("git", "clone", "https://github.com/wolfogre/st.git", REPO_PATH).Run()
		if err != nil {
			log.Panic(err)
		}
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
	cmd.Dir = REPO_PATH
	err := cmd.Run()
	if err != nil {
		status = err.Error()
	} else {
		status = ""
	}
	c.JSON(http.StatusOK, gin.H{
		"msg": "OK",
	})

}

func handleIndex(c *gin.Context) {
	c.File(REPO_PATH + "/index.sh")
}

func handleFunc(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	if _, err := os.Stat(REPO_PATH + "/func/" + name); os.IsNotExist(err) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.File(REPO_PATH + "/func/" + name)
}