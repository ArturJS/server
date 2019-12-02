package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.POST("/deploy", func(c *gin.Context) {
		file, err := c.FormFile("file")

		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		err = c.SaveUploadedFile(
			file,
			fmt.Sprintf("/home/serve/deployments/%d", time.Now().Unix()),
		)

		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		c.JSON(http.StatusOK, "done")
	})

	log.Fatal(router.Run(":8080"))
}
