package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

//Page
type Page struct {
	Title   string
	Link    string
	Snippet string
}

//Query
type Query struct {
	Items []Page
}

var myClient = &http.Client{Timeout: 10 * time.Second}
var userQuery = "Perritos"

func getJSON(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func main() {
	var apiURL = "https://www.googleapis.com/customsearch/v1?q=QUERY&cx=006924283690115384884%3Aci07khskaey&key=AIzaSyDnjzE_wsZ7Bo2KekvjnDvTgZFFLkezhT4"
	r := gin.Default()
	r.GET("/:query", func(c *gin.Context) {
		query := new(Query) // or &Foo{}
		getJSON(strings.Replace(apiURL, "QUERY", c.Param("query"), 1), query)
		c.JSON(200, query)
	})
	r.Run() // listen and serve on 0.0.0.0:8080apiURL = strings.Replace(apiURL, "QUERY", userQuery, 1)

}
