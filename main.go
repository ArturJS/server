package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mholt/archiver/v3"
)

func main() {
	var (
		logs []string
		zip  archiver.Zip
	)

	router := gin.Default()
	router.MaxMultipartMemory = 32 << 20 // 8 MiB
	router.POST("/deploy", func(c *gin.Context) {
		timestamp := time.Now().Unix()
		service := c.PostForm("service")
		file, err := c.FormFile("file")

		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		// create deployment dir
		err = os.MkdirAll(
			fmt.Sprintf("/home/serve/deployments/%s", service),
			os.ModePerm,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		// save file
		err = c.SaveUploadedFile(
			file,
			fmt.Sprintf("/home/serve/deployments/%s/%d.zip", service, timestamp),
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		// create build dir
		err = os.MkdirAll(
			fmt.Sprintf("/home/serve/builds/%s/%d", service, timestamp),
			os.ModePerm,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		// unzip to build dir
		err = zip.Unarchive(
			fmt.Sprintf("/home/serve/deployments/%s/%d.zip", service, timestamp),
			fmt.Sprintf("/home/serve/builds/%s/%d", service, timestamp),
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		// create container dir
		err = os.MkdirAll(
			fmt.Sprintf("/home/serve/containers/%s", service),
			os.ModePerm,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		// remove symlink if exists
		symlink := fmt.Sprintf("/home/serve/containers/%s/latest", service)

		if _, err := os.Lstat(symlink); err == nil {
			if err = os.Remove(symlink); err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
		}

		// symlink
		err = os.Symlink(
			fmt.Sprintf("/home/serve/builds/%s/%d", service, timestamp),
			symlink,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		// start docker container
		ctx := context.Background()
		dockerCmd := exec.CommandContext(
			ctx,
			"/usr/local/bin/docker-compose",
			"-p serve",
			"-f",
			fmt.Sprintf("/home/serve/containers/%s/latest/docker-compose.yml", service),
			"up",
			"-d",
			"--build",
			service,
		)

		dockerCmd.Dir = symlink

		out, err := dockerCmd.Output()

		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"out": out,
				"err": err,
			})
			return
		}

		// done
		c.JSON(http.StatusOK, map[string]interface{}{
			"logs": logs,
			"out":  string(out),
		})
	})

	log.Fatal(router.Run(":8080"))
}
