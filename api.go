package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mholt/archiver/v3"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"
)

type api struct {
	zip  *archiver.Zip
	port string
	mode string
	*sync.RWMutex
}

func (api *api) getMode() string {
	api.RLock()
	defer api.RUnlock()

	return api.mode
}

func (api *api) getPort() string {
	api.RLock()
	defer api.RUnlock()

	return api.port
}

func (api *api) getZip() *archiver.Zip {
	api.RLock()
	defer api.RUnlock()

	return api.zip
}

func (api *api) response(error, data interface{}) gin.H {
	return gin.H{
		"error": error,
		"data":  data,
	}
}

func (api *api) startServer() error {
	gin.SetMode(api.getMode())
	r := gin.Default()

	// routes
	r.GET("/ok", api.ok)
	r.POST("/deploy", api.deploy)

	return r.Run(":" + api.getPort())
}

func (api *api) ok(c *gin.Context) {
	c.JSON(http.StatusOK, api.response(nil, "ok"))
}

func (api *api) deploy(c *gin.Context) {
	timestamp := time.Now().Unix()
	service := c.PostForm("service")
	file, err := c.FormFile("file")

	if err != nil {
		c.JSON(http.StatusInternalServerError, api.response(err.Error(), nil))
		return
	}

	// create deployment dir
	err = os.MkdirAll(
		fmt.Sprintf("/home/serve/deployments/%s", service),
		os.ModePerm,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, api.response(err.Error(), nil))
		return
	}

	// save file
	err = c.SaveUploadedFile(
		file,
		fmt.Sprintf("/home/serve/deployments/%s/%d.zip", service, timestamp),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, api.response(err.Error(), nil))
		return
	}

	// create build dir
	err = os.MkdirAll(
		fmt.Sprintf("/home/serve/builds/%s/%d", service, timestamp),
		os.ModePerm,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, api.response(err.Error(), nil))
		return
	}

	// unzip to build dir
	err = api.getZip().Unarchive(
		fmt.Sprintf("/home/serve/deployments/%s/%d.zip", service, timestamp),
		fmt.Sprintf("/home/serve/builds/%s/%d", service, timestamp),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, api.response(err.Error(), nil))
		return
	}

	// create container dir
	err = os.MkdirAll(
		fmt.Sprintf("/home/serve/containers/%s", service),
		os.ModePerm,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, api.response(err.Error(), nil))
		return
	}

	// remove symlink if exists
	symlink := fmt.Sprintf("/home/serve/containers/%s/latest", service)

	if _, err := os.Lstat(symlink); err == nil {
		if err = os.Remove(symlink); err != nil {
			c.JSON(http.StatusInternalServerError, api.response(err.Error(), nil))
			return
		}
	}

	// symlink
	err = os.Symlink(
		fmt.Sprintf("/home/serve/builds/%s/%d", service, timestamp),
		symlink,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, api.response(err.Error(), nil))
		return
	}

	// start docker container
	ctx := context.Background()
	dockerCmd := exec.CommandContext(
		ctx,
		"/usr/local/bin/docker-compose",
		"-p serve",
		"-f",
		"/home/serve/containers/*/latest/docker-compose.yml",
		"up",
		"-d",
		"--build",
		service,
	)

	dockerCmd.Dir = symlink

	// output
	out, err := dockerCmd.Output()

	if err != nil {
		c.JSON(http.StatusInternalServerError, api.response(err.Error(), out))
		return
	}

	// done
	c.JSON(http.StatusOK, api.response(nil, "done"))
}