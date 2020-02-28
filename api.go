package main

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/mholt/archiver/v3"
	"github.com/otiai10/copy"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type api struct {
	ssl      bool
	zip      *archiver.Zip
	port     string
	mode     string
	maxSize  int64
	token    string
	certFile string
	certKey  string

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

func (api *api) response(error, data interface{}, base64 bool) gin.H {
	return gin.H{
		"error":  error,
		"data":   data,
		"base64": base64,
	}
}

func (api *api) authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader := c.Request.Header.Get("Authorization")

		if authorizationHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.response("authorization header is missing", nil, false))
			return
		}

		bearerToken := strings.Split(authorizationHeader, " ")

		if len(bearerToken) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.response("invalid bearer token", nil, false))
			return
		}

		token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(api.getToken()), nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.response(err.Error(), nil, false))
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.response("invalid token", nil, false))
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

	if api.ssl {
		return r.RunTLS(":"+api.getPort(), certFile, keyFile)
	}

	return r.Run(":" + api.getPort())
}

func (api *api) ok(c *gin.Context) {
	c.JSON(http.StatusOK, api.response(nil, "ok", false))
}

func (api *api) deploy(c *gin.Context) {
	var config config

	ctx := context.Background()
	timestamp := time.Now().Unix()

	// config
	err := yaml.Unmarshal([]byte(c.PostForm("config")), &config)

	if err != nil {
		c.JSON(http.StatusInternalServerError, api.response(err.Error(), nil, false))
		return
	}

	// file
	file, err := c.FormFile("file")

	if err != nil {
		c.JSON(http.StatusInternalServerError, api.response(err.Error(), nil, false))
		return
	}

	// shared
	isShared := len(config.Service) == 0

	// create deployment dir
	err = os.MkdirAll(
		fmt.Sprintf("/home/makeless/deployments/%s", config.Name),
		os.ModePerm,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, api.response(err.Error(), nil, false))
		return
	}

	// save file
	err = c.SaveUploadedFile(
		file,
		fmt.Sprintf("/home/makeless/deployments/%s/%d.zip", config.Name, timestamp),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, api.response(err.Error(), nil, false))
		return
	}

	// create build dir
	err = os.MkdirAll(
		fmt.Sprintf("/home/makeless/builds/%s/%d", config.Name, timestamp),
		os.ModePerm,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, api.response(err.Error(), nil, false))
		return
	}

	// unzip to build dir
	err = api.getZip().Unarchive(
		fmt.Sprintf("/home/makeless/deployments/%s/%d.zip", config.Name, timestamp),
		fmt.Sprintf("/home/makeless/builds/%s/%d", config.Name, timestamp),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, api.response(err.Error(), nil, false))
		return
	}

	// create container dir
	err = os.MkdirAll(
		fmt.Sprintf("/home/makeless/containers/%s", config.Name),
		os.ModePerm,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, api.response(err.Error(), nil, false))
		return
	}

	// remove symlink if exists
	symlink := fmt.Sprintf("/home/makeless/containers/%s/latest", config.Name)

	if _, err := os.Lstat(symlink); err == nil {
		if err = os.Remove(symlink); err != nil {
			c.JSON(http.StatusInternalServerError, api.response(err.Error(), nil, false))
			return
		}
	}

	// symlink
	err = os.Symlink(
		fmt.Sprintf("/home/makeless/builds/%s/%d", config.Name, timestamp),
		symlink,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, api.response(err.Error(), nil, false))
		return
	}

	// shared
	if isShared {
		c.JSON(http.StatusOK, api.response(nil, "done", false))
		return
	}

	// use
	for service, items := range config.Use {
		for _, path := range items {
			split := strings.Split(path, ":")

			src := fmt.Sprintf("/home/makeless/containers/%s/latest/%s", service, split[0])
			dst := fmt.Sprintf("/home/makeless/containers/%s/latest/%s", config.Name, split[1])

			if err := copy.Copy(src, dst); err != nil {
				c.JSON(http.StatusInternalServerError, api.response(err.Error(), nil, false))
				return
			}
		}
	}

	// start docker container
	// build args
	args := new(args)
	args.push("-p", "makeless")

	// push all containers
	containers, err := ioutil.ReadDir("/home/makeless/containers")

	if err != nil {
		c.JSON(http.StatusInternalServerError, api.response(err.Error(), nil, false))
		return
	}

	for _, container := range containers {
		if !container.IsDir() {
			continue
		}

		args.push(
			"-f",
			fmt.Sprintf("/home/makeless/containers/%s/latest/docker-compose.yml", container.Name()),
		)
	}

	args.push("up", "-d", "--build", config.Name)

	// docker command
	dockerCmd := exec.CommandContext(ctx, "/usr/local/bin/docker-compose", args.data...)
	dockerCmd.Dir = symlink

	// output
	out, err := dockerCmd.CombinedOutput()

	if err != nil {
		c.JSON(http.StatusInternalServerError, api.response(err.Error(), out, true))
		return
	}

	// done
	c.JSON(http.StatusOK, api.response(nil, out, true))
}
