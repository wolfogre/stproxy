package main

import (
	"os"
	"os/exec"
	"log"
	"github.com/gin-gonic/gin"
	"net/http"
	"io/ioutil"
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
	"encoding/json"
	"time"
	"strings"
)

const (
	REPO_PATH = "/opt/st"
)

var (
	status = ""
	secret = ""
)

func main() {
	log.SetFlags(log.LstdFlags | log.LstdFlags)
	if _, err := os.Stat(REPO_PATH); os.IsNotExist(err) {
		err := exec.Command("git", "clone", "https://github.com/wolfogre/st.git", REPO_PATH).Run()
		if err != nil {
			log.Panic(err)
		}
	}

	secret = os.Getenv("SECRET")
	log.Println("secret: ", secret)

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
	buffer, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if secret != "" {
		h := hmac.New(sha1.New, []byte(secret))
		n, err := h.Write(buffer)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if n != len(buffer) {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		signature := fmt.Sprintf("sha1=%x", h.Sum(nil))
		if signature != c.GetHeader("X-Hub-Signature") {
			c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("wrong signature: %v", c.GetHeader("X-Hub-Signature")))
			return
		}
	}

	if c.GetHeader("X-Github-Event") != "push" {
		c.JSON(http.StatusOK, gin.H{
			"msg": "not push event",
		})
		return
	}

	body := make(map[string]interface{})
	err = json.Unmarshal(buffer, body)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	go pull(body["after"].(string))

	c.JSON(http.StatusOK, gin.H{
		"msg": "OK",
	})

}

func pull(commit string) {
	for i := 0; i < 60; i++ {
		time.Sleep(time.Second)
		log.Println("git pull ", commit)
		cmd := exec.Command("git", "pull")
		cmd.Dir = REPO_PATH
		err := cmd.Run()
		if err != nil {
			log.Println(err)
			status = err.Error()
			continue
		}
		buffer, err := ioutil.ReadFile(REPO_PATH + "/.git/refs/heads/master")
		if err != nil {
			log.Println(err)
			continue
		}
		lastCommit := strings.TrimSpace(string(buffer))
		log.Println("last commit: ", lastCommit)
		if lastCommit == commit {
			break
		}
	}
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