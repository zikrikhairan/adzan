package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*.html")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	r.GET("/prayer", func(c *gin.Context) {

		// Let's first read the `config.json` file
		content, err := ioutil.ReadFile("data/mosque_0001.json")
		if err != nil {
			log.Fatal("Error when opening file: ", err)
		}

		// Now let's unmarshall the data into `payload`
		var response []interface{}
		err = json.Unmarshal(content, &response)
		if err != nil {
			log.Fatal("Error during Unmarshal(): ", err)
		}
		c.JSON(http.StatusOK, response)
	})
	r.Run("127.0.0.1:53404")
}
