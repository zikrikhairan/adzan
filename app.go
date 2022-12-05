package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
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
	r.GET("/getAllMosque", func(c *gin.Context) {
		getAllMosque()
		c.JSON(http.StatusOK, "Successfully retrieved all")
	})

	r.GET("/removeDuplicateMosque", func(c *gin.Context) {
		removeDuplicateMosque()
		c.JSON(http.StatusOK, "Successfully retrieved all")
	})

	r.GET("/getPrayerTimeByLocation", func(c *gin.Context) {
		const bitSize = 64 // Don't think about it to much. It's just 64 bits.
		latitudeQuery := c.Query("latitude")
		longitudeQuery := c.Query("longitude")
		monthQuery := c.Query("month")
		yearQuery := c.Query("year")
		latitude, _ := strconv.ParseFloat(latitudeQuery, bitSize)
		longitude, _ := strconv.ParseFloat(longitudeQuery, bitSize)
		month, _ := strconv.Atoi(monthQuery)
		year, _ := strconv.Atoi(yearQuery)
		response, _ := getPrayerTimeByLocation(latitude, longitude, month, year)
		c.JSON(http.StatusOK, response)
	})

	r.GET("/getPrayerTimeByMosqueLocation", func(c *gin.Context) {
		monthQuery := c.Query("month")
		yearQuery := c.Query("year")
		month, _ := strconv.Atoi(monthQuery)
		year, _ := strconv.Atoi(yearQuery)
		getPrayerTimeByMosqueLocation(month, year)
		c.JSON(http.StatusOK, "Successfully retrieved all")
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

type Mosque struct {
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
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

func getAllMosque() {
	start := time.Now()
	defer func() {
		fmt.Println("Execution Time: ", time.Since(start))
	}()
	//minLat := -90
	//maxLat := 90
	//minLong := -180
	//maxLong := 180

	minLat := -180
	maxLat := 180
	minLong := -90
	maxLong := 90

	//with concurrent
	//for x := minLat; x <= maxLat; x++ {
	//	wg := sync.WaitGroup{}
	//	var listMosque []Candidate
	//	for y := minLong; y <= maxLong; y++ {
	//		wg.Add(1)
	//		go func(x int, y int) {
	//			dataMosque, err := getMosqueLocation(float64(x), float64(y))
	//			if err != nil {
	//				return
	//			}
	//			for _, mosque := range dataMosque.Candidates {
	//				listMosque = append(listMosque, mosque)
	//			}
	//			wg.Done()
	//		}(x, y)
	//	}
	//	wg.Wait()
	//	file, _ := json.MarshalIndent(listMosque, "", " ")
	//	fileName := fmt.Sprintf("%s.json", strconv.Itoa(x))
	//	_ = ioutil.WriteFile("data/mosque/"+fileName, file, 0644)
	//}

	//without concurrent
	for x := minLat; x <= maxLat; x++ {
		var listMosque []Candidate
		for y := minLong; y <= maxLong; y++ {
			dataMosque, err := getMosqueLocation(float64(x), float64(y))
			if err != nil {
				return
			}
			for _, mosque := range dataMosque.Candidates {
				listMosque = append(listMosque, mosque)
			}
		}
		file, _ := json.MarshalIndent(listMosque, "", " ")
		fileName := fmt.Sprintf("%s.json", strconv.Itoa(x))
		_ = ioutil.WriteFile("data/mosque/"+fileName, file, 0644)
	}
}

func removeDuplicateMosque() {
	path := "data/mosque/"
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	listKeyMosque := make(map[string]bool)
	var listAllMosque []Mosque
	for _, file := range files {
		fmt.Println(file.Name(), file.IsDir())
		content, err := ioutil.ReadFile(path + file.Name())
		if err != nil {
			log.Fatal("Error when opening file: ", err)
		}

		// Now let's unmarshall the data into `payload`
		var listMosque []Candidate
		err = json.Unmarshal(content, &listMosque)
		if err != nil {
			log.Fatal("Error during Unmarshal(): ", err)
		}
		for _, candidate := range listMosque {
			latitude := strconv.FormatFloat(candidate.Location.Latitude, 'E', -1, 32)
			longitude := strconv.FormatFloat(candidate.Location.Longitude, 'E', -1, 32)
			key := fmt.Sprintf("%s,%s", latitude, longitude)
			if _, ok := listKeyMosque[key]; ok {
				continue
			} else {
				listKeyMosque[key] = true
				var mosque Mosque
				mosque.Name = candidate.Attribute.Name
				mosque.Country = candidate.Attribute.Country
				mosque.Latitude = candidate.Location.Latitude
				mosque.Longitude = candidate.Location.Longitude
				listAllMosque = append(listAllMosque, mosque)
			}
		}
	}

	file, _ := json.MarshalIndent(listAllMosque, "", " ")
	fileName := "mosque.json"
	_ = ioutil.WriteFile("data/"+fileName, file, 0644)
}

type ResponsePrayer struct {
	Data []Prayer `json:"data"`
}

type Prayer struct {
	Timings Timings `json:"timings"`
	Date    Date    `json:"date"`
}

type Timings struct {
	Fajr    string `json:"Fajr"`
	Dhuhr   string `json:"Dhuhr"`
	Asr     string `json:"Asr"`
	Maghrib string `json:"Maghrib"`
	Isha    string `json:"Isha"`
}

type Date struct {
	Readable  string `json:"readable"`
	Timestamp string `json:"timestamp"`
}

type PrayerTime struct {
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Shalat    string  `json:"shalat"`
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

const baseURLPrayer = "http://api.aladhan.com/v1/calendar?latitude=%f&longitude=%f&month=%d&year=%d&iso8601=true&timezonestring=UTC&method=2"

func getPrayerTimeByLocation(lat float64, long float64, month int, year int) (responsePrayer *ResponsePrayer, err error) {
	data := url.Values{}
	urlStr := fmt.Sprintf(baseURLPrayer, lat, long, month, year)

	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodGet, urlStr, strings.NewReader(data.Encode())) // URL-encoded payload

	response, _ := client.Do(r)
	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(response.Body).Decode(&responsePrayer)
	if err != nil {
		return nil, err
	}

	return responsePrayer, nil
}

const dbMaxIdleConns = 4
const dbMaxConns = 100
const totalWorker = 100

func openDbConnection() (*sql.DB, error) {
	host := os.Getenv("PG_HOSTNAME")
	port, _ := strconv.Atoi(os.Getenv("PG_PORT"))
	user := os.Getenv("PG_USERNAME")
	password := os.Getenv("PG_PASSWORD")
	dbname := os.Getenv("PG_DATABASE")
	// connection string
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlconn)
	CheckError(err)
	db.SetMaxOpenConns(dbMaxConns)
	db.SetMaxIdleConns(totalWorker)
	return db, nil
}

func getPrayerTimeByMosqueLocation(month int, year int) {
	db, _ := openDbConnection()
	content, err := ioutil.ReadFile("data/mosque.json")
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}
	start := time.Now()
	defer func() {
		fmt.Println("Execution Time: ", time.Since(start))
	}()
	//Truncate table prayer time
	truncateTable := `truncate table "prayer_time"`
	_, e := db.Exec(truncateTable)
	CheckError(e)

	// Now let's unmarshall the data into `payload`
	var listMosque []Mosque
	err = json.Unmarshal(content, &listMosque)

	fmt.Println(len(listMosque))
	for index, mosque := range listMosque {
		response, _ := getPrayerTimeByLocation(mosque.Latitude, mosque.Longitude, month, year)
		for _, dataPrayer := range response.Data {
			go func() {
				fajr := dataPrayer.Timings.Fajr
				dhuhr := dataPrayer.Timings.Dhuhr
				asr := dataPrayer.Timings.Asr
				maghrib := dataPrayer.Timings.Maghrib
				isha := dataPrayer.Timings.Isha
				fajrDate, _ := time.Parse(time.RFC3339, strings.Replace(fajr, "+00:00 (UTC)", "", 1)+"Z")
				dhuhrDate, _ := time.Parse(time.RFC3339, strings.Replace(dhuhr, "+00:00 (UTC)", "", 1)+"Z")
				asrDate, _ := time.Parse(time.RFC3339, strings.Replace(asr, "+00:00 (UTC)", "", 1)+"Z")
				maghribDate, _ := time.Parse(time.RFC3339, strings.Replace(maghrib, "+00:00 (UTC)", "", 1)+"Z")
				ishaDate, _ := time.Parse(time.RFC3339, strings.Replace(isha, "+00:00 (UTC)", "", 1)+"Z")
				fajrEpoch := fajrDate.Unix()
				dhuhrEpoch := dhuhrDate.Unix()
				asrEpoch := asrDate.Unix()
				maghribEpoch := maghribDate.Unix()
				ishaEpoch := ishaDate.Unix()
				checkToListEpoch(mosque, "Fajr", fajrEpoch, db)
				checkToListEpoch(mosque, "Dhuhr", dhuhrEpoch, db)
				checkToListEpoch(mosque, "Asr", asrEpoch, db)
				checkToListEpoch(mosque, "maghrib", maghribEpoch, db)
				checkToListEpoch(mosque, "Isha", ishaEpoch, db)
			}()
		}
		fmt.Println(index)
	}
}

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func checkToListEpoch(mosque Mosque, shalat string, epoch int64, database *sql.DB) {
	insertDynStmt := `insert into "prayer_time"("time", "name", "country", "shalat", "lat", "lon") values($1, $2, $3, $4, $5, $6)`
	_, e := database.Exec(insertDynStmt, epoch, mosque.Name, mosque.Country, shalat, mosque.Latitude, mosque.Longitude)
	CheckError(e)
}
