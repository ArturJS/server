package main

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/mholt/archiver/v3"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type api struct {
	zip     *archiver.Zip
	port    string
	mode    string
	maxSize int64
	token   string

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

func (api *api) getMaxSize() int64 {
	api.RLock()
	defer api.RUnlock()

	return api.maxSize
}

func (api *api) getToken() string {
	api.RLock()
	defer api.RUnlock()

	return api.token
}

func (api *api) response(error, data interface{}) gin.H {
	return gin.H{
		"error": error,
		"data":  data,
	}
}

func (api *api) authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader := c.Request.Header.Get("Authorization")

		if authorizationHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.response("authorization header is missing", nil))
			return
		}

		bearerToken := strings.Split(authorizationHeader, " ")

		if len(bearerToken) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.response("invalid bearer token", nil))
			return
		}

		token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(api.getToken()), nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.response(err.Error(), nil))
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.response("invalid token", nil))
			return
		}
	}
}

func (api *api) startServer() error {
	gin.SetMode(api.getMode())
	r := gin.Default()

	// config
	r.MaxMultipartMemory = api.getMaxSize() << 20

	// routes
	r.GET("/ok", api.ok)

	// auth routes
	authGroup := r.Group("/")
	authGroup.Use(api.authenticate())
	{
		authGroup.POST("/deploy", api.deploy)
	}

	return r.Run(":" + api.getPort())
}

func (api *api) ok(c *gin.Context) {
	c.JSON(http.StatusOK, api.response(nil, "ok"))
}

func (api *api) deploy(c *gin.Context) {
	ctx := context.Background()
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
	// build args
	args := new(args)
	args.push("-p", "serve")

	// push all containers
	containers, err := ioutil.ReadDir("/home/serve/containers")

	if err != nil {
		c.JSON(http.StatusInternalServerError, api.response(err.Error(), nil))
		return
	}

	for _, container := range containers {
		if !container.IsDir() {
			continue
		}

		args.push(
			"-f",
			fmt.Sprintf("/home/serve/containers/%s/latest/docker-compose.yml", container.Name()),
		)
	}

	args.push("up", "-d", "--build", service)

	// docker command
	dockerCmd := exec.CommandContext(ctx, "/usr/local/bin/docker-compose", args.data...)
	dockerCmd.Dir = symlink

	// output
	out, err := dockerCmd.CombinedOutput()

	if err != nil {
		c.JSON(http.StatusInternalServerError, api.response(err.Error(), out))
		return
	}

	// done
	c.JSON(http.StatusOK, api.response(nil, "done"))
}
