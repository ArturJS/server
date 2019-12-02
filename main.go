package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mholt/archiver/v3"
)

func main() {
	zip := archiver.Zip{}
	router := gin.Default()
	router.MaxMultipartMemory = 32 << 20 // 8 MiB
	router.POST("/deploy", func(c *gin.Context) {
		timestamp := time.Now().Unix()
		service := c.PostForm("service")
		zipName := fmt.Sprintf("%s-%d.zip", service, timestamp)

		file, err := c.FormFile("file")

		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		// save file
		err = c.SaveUploadedFile(
			file,
			fmt.Sprintf("/home/serve/deployments/%s", zipName),
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
			fmt.Sprintf("/home/serve/deployments/%s", zipName),
			fmt.Sprintf("/home/serve/builds/%s/%d", service, timestamp),
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, "done")
	})

	log.Fatal(router.Run(":8080"))
}
