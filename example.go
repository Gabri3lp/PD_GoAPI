package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gin-gonic/gin"
)

//The Page struct is the representation of the page json returned by the api.
type Page struct {
	Id            int
	Title         string
	Link          string
	Snippet       string
	IsHighlighted bool
}

//The QueryResponse struct is an aray of Items which are pages.
type QueryResponse struct {
	Items []Page
}

type UClassifyResponse struct {
	Computers float64
	Arts      float64
}

var myClient = &http.Client{Timeout: 10 * time.Second}

const GOOGLE_API_URL = "https://www.googleapis.com/customsearch/v1"
const GOOGLE_API_KEY = "AIzaSyCy_VyBA9JVye7KnHVYS0vJxnMAutMUNAQ"
const GOOGLE_API_CX = "001983809218396823816:rnaroujms5e"
const FILTER_CATEGORY = "technology"

const UCLASSIFY_API = "https://api.uclassify.com/v1/uClassify/Topics/classify/"
const UCLASSIFY_KEY = "C904XyojCY61"

var apiURL = "https://www.googleapis.com/customsearch/v1?key=AIzaSyCbF2sNyXVkLMVN_5T0yWaFNAUYTdhUz-8&cx=001983809218396823816:rnaroujms5e&q=QUERY"

func main() {

	db, err := sql.Open("mysql", "root@tcp(localhost:3306)/dist")
	if err != nil {
		panic(err.Error())
	}
	r := gin.New()
	r.Use(CORSMiddleware())
	r.GET("/web/:query", func(c *gin.Context) {
		req, _ := http.NewRequest("GET", GOOGLE_API_URL, nil)
		q := req.URL.Query()
		q.Set("key", GOOGLE_API_KEY)
		q.Add("cx", GOOGLE_API_CX)
		q.Add("fields", "items(title,snippet,link)")
		q.Add("q", c.Param("query"))
		req.URL.RawQuery = q.Encode()
		response := new(QueryResponse)
		getJSON(req, response)
		result := new(QueryResponse)
		for _, page := range response.Items {
			categories := getCategory(page.Snippet)
			if categories == "technology" {
				result.Items = append(result.Items, page)
				fmt.Println(page.Link)
			}
		}
		checkHighlight(c.Param("query"), result, db)
		c.JSON(200, result.Items)

	})
	r.GET("/images/:query", func(c *gin.Context) {
		req, _ := http.NewRequest("GET", GOOGLE_API_URL, nil)
		q := req.URL.Query()
		q.Set("key", GOOGLE_API_KEY)
		q.Add("cx", GOOGLE_API_CX)
		q.Add("fields", "items(title,snippet,link)")
		q.Add("searchType", "image")
		q.Add("q", c.Param("query"))
		req.URL.RawQuery = q.Encode()
		response := new(QueryResponse)
		getJSON(req, response)
		result := new(QueryResponse)
		for _, page := range response.Items {
			categories := getCategory(page.Snippet)
			if categories == FILTER_CATEGORY {
				result.Items = append(result.Items, page)
				fmt.Println(page.Link)
			}
		}
		checkHighlight(c.Param("query"), result, db)
		c.JSON(200, result.Items)
	})
	r.GET("/highlight", func(c *gin.Context) {
		highlight(c, db)
		c.JSON(200, gin.H{"result": true})
	})
	r.Run() // listen and serve on 0.0.0.0:8080

}
func getCategory(text string) string {
	req, err := http.NewRequest("GET", UCLASSIFY_API, nil)
	q := req.URL.Query()
	q.Add("readKey", UCLASSIFY_KEY)
	q.Add("text", text)
	req.URL.RawQuery = q.Encode()
	resp, err := myClient.Do(req)
	if err != nil || resp == nil {
		return ""
	}
	response := new(UClassifyResponse)
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	resp.Body.Close()
	if response.Computers >= 0.5 {
		return FILTER_CATEGORY
	}
	return ""
}
func checkHighlight(query string, result *QueryResponse, db *sql.DB) {
	for i := 0; i < len(result.Items); i++ {
		result.Items[i].IsHighlighted = false
	}
	rows, err := db.Query("SELECT p.id_pagina, p.titulo, p.descripcion, p.url FROM resalta r,  busqueda b, pagina p WHERE b.id_busqueda = r.id_busqueda && p.id_pagina = r.id_pagina && b.palabra = ?", query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var page *Page
	for rows.Next() {
		page = new(Page)
		err := rows.Scan(&page.Id, &page.Title, &page.Snippet, &page.Link)
		if err != nil {
			log.Fatal(err)
		}
		result.Items = append([]Page{*page}, result.Items...)
		result.Items[0].IsHighlighted = true
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}
func highlight(c *gin.Context, db *sql.DB) {
	if c.Query("Id") == "-1" {
		page := Page{Title: c.Query("Title"), Snippet: c.Query("Snippet"), Link: c.Query("Link")}
		query := c.Query("Query")
		rows, err := db.Query("CALL Highlight(?, ?, ?, ?)", query, page.Title, page.Snippet, page.Link)
		defer rows.Close()
		if err != nil {
			panic(err.Error())
		}
	} else {
		rows, err := db.Query("CALL Remove_Highlight(?)", c.Query("Id"))
		defer rows.Close()
		if err != nil {
			panic(err.Error())
		}
	}

}

//CORSMiddleware ,Middleware CORS Handler
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "*")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func getJSON(r *http.Request, target interface{}) error {
	resp, err := myClient.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}
