package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error read .env: ", err)
	}
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
	r.GET("/getMosqueByPosition", func(c *gin.Context) {
		const bitSize = 64 // Don't think about it to much. It's just 64 bits.
		latitudeQuery := c.Query("latitude")
		longitudeQuery := c.Query("longitude")
		latitude, _ := strconv.ParseFloat(latitudeQuery, bitSize)
		longitude, _ := strconv.ParseFloat(longitudeQuery, bitSize)
		response, _ := getMosqueLocation(latitude, longitude)
		c.JSON(http.StatusOK, response)
	})
	r.Run("127.0.0.1:53404")
}

type ResponseSpatial struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Attribute Attribute `json:"attributes"`
	Location  Location  `json:"location"`
}

type Attribute struct {
	Name    string `json:"PlaceName"`
	Address string `json:"Place_addr"`
	Country string `json:"Country"`
}

type Location struct {
	Latitude  float64 `json:"y"`
	Longitude float64 `json:"x"`
}

const baseURL = "https://geocode-api.arcgis.com/"
const path = "arcgis/rest/services/World/GeocodeServer/findAddressCandidates"

func getMosqueLocation(lat float64, long float64) (responseSpatial *ResponseSpatial, err error) {
	data := url.Values{}
	data.Set("f", "json")
	data.Set("category", "Mosque")
	data.Set("location", fmt.Sprintf("%f,%f", lat, long))
	data.Set("maxLocations", "200")
	data.Set("outFields", "Place_addr PlaceName Country")
	token := os.Getenv("ARCGIS_TOKEN")
	data.Set("token", token)

	u, _ := url.ParseRequestURI(baseURL)
	u.Path = path
	urlStr := u.String()

	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(data.Encode())) // URL-encoded payload
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, _ := client.Do(r)
	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(response.Body).Decode(&responseSpatial)
	if err != nil {
		return nil, err
	}

	return responseSpatial, nil
}
