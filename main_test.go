package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
)

const PostgresConn = "host=localhost port=5432 user=postgres password=mysecretpassword dbname=postgres sslmode=disable"
const MySQLConn = "root:my-secret-pw@tcp(localhost:3306)/db"

func TestConnectToDB(t *testing.T) {

	logFile, _ := os.OpenFile(
		"./restful_api.log",
		os.O_TRUNC|os.O_CREATE|os.O_RDWR,
		os.ModePerm,
	)

	logger := log.New(logFile, "INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	// positive case for mysql DB conn
	dbConn, err := connectToDB("mysql", MySQLConn, logger)
	if err != nil || dbConn.Ping() != nil {
		t.Errorf("Failed to establish mysql DB connection")
	}
	dbConn.Close()

	// negative case for mysql DB conn
	dbConn, err = connectToDB("mysql", "unknown:unknown@tcp(localhost:3306)/db", logger)
	if err == nil || dbConn.Ping() == nil {
		t.Errorf("DB Connection established for wrong user name / password")
		dbConn.Close()
	}

	// negative case for mysql DB conn
	dbConn, err = connectToDB("mysql", "myuser:mypass@tcp(localhost:1234)/db", logger)
	if err == nil || dbConn.Ping() == nil {
		t.Errorf("DB Connection established for db host/port")
	}
	dbConn.Close()

	// positive case for postgres DB conn
	dbConn, err = connectToDB("postgres", PostgresConn, logger)
	if err != nil || dbConn.Ping() != nil {
		t.Errorf("Failed to establish mysql DB connection")
	}
	dbConn.Close()

	// negative case for postgres DB conn
	dbConn, err = connectToDB("postgres", "host=localhost port=5432 user=unknown password=unknown dbname=postgres sslmode=disable", logger)
	if err == nil || dbConn.Ping() == nil {
		t.Errorf("DB Connection established for wrong user name / password")
	}
	dbConn.Close()

	// negative case for postgres DB conn
	dbConn, err = connectToDB("postgres", "host=localhost port=1234 user=postgres password=mysecretpassword dbname=postgres sslmode=disable", logger)
	if err == nil || dbConn.Ping() == nil {
		t.Errorf("DB Connection established for db host/port")
	}
	dbConn.Close()
}

func initTestModule(db, connectionString string) (app *App) {

	logFile, _ := os.OpenFile(
		"./restful_api.log",
		os.O_TRUNC|os.O_CREATE|os.O_RDWR,
		os.ModePerm,
	)

	logger := log.New(logFile, "INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	// connect to DB
	dbConn, err := connectToDB(db, connectionString, logger)
	if err != nil {
		log.Println(err)
	}

	// set new router
	app = &App{
		DBType:   db,
		Router:   mux.NewRouter().StrictSlash(true),
		Database: dbConn,
		logger:   logger,
	}

	return
}

func TestHomepage(t *testing.T) {

	const homepageResponse = `
- POST /article
  - Add new article to DB
  - payload :
    {
        Title     (string)
        desc      (string)
        content   (string)
    }

- PUT /article/{id}
  - Update an existing article DB
  - query param : id (article id from GET API)
  - payload :
    {
        Title     (string)
        desc      (string)
        content   (string)
    }

- DELETE /article/{id}
  - Deletes an entry from DB
  - query param : id (article id from GET API)

- GET /article/{id}
  - Retrieves article data from DB for a given ID
  - query param : id (article id from GET API) 

- GET /articles
  - retrives all articles from DB
  - query params : id (last ID from previous GET call for pagination), limit (max entry per page)
  - response : list of articles
`

	// creating new request to homepage
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// start the DB connection and router for app
	app := initTestModule("mysql", MySQLConn)

	// if db connection fails, add as fatal error
	if app.Database == nil || app.Database.Ping() != nil {
		t.Fatal(err)
	}

	// new recorder for capturing response from request
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.homepage)

	// serve http call on request
	handler.ServeHTTP(rr, req)

	// checking http status code if 200
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response string matches expected response
	if rr.Body.String() != homepageResponse {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), homepageResponse)
	}
}

func TestCreateNewArticle(t *testing.T) {

	for _, dbType := range []string{"mysql", "postgres"} {

		var (
			connectionString string
			responseArticle  Article
		)

		if dbType == "mysql" {
			connectionString = MySQLConn
		} else {
			connectionString = PostgresConn
		}

		// start the DB connection and router for app
		app := initTestModule(
			dbType,
			connectionString,
		)

		// prepare article data
		article := Article{
			Title:   "TestArticle",
			Desc:    "TestDesc",
			Content: "TestContent",
		}

		// convert the article data as json
		payload, err := json.Marshal(article)
		if err != nil {
			t.Fatal(err)
		}

		// creating new request to homepage
		req, err := http.NewRequest("POST", "/article", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatal(err)
		}

		// if db connection fails, add as fatal error
		if app.Database == nil || app.Database.Ping() != nil {
			t.Fatal(err)
		}

		// new recorder for capturing response from request
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(app.createNewArticle)

		// serve http call on request
		handler.ServeHTTP(rr, req)

		// checking http status code if 200
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		// decode the response body to a new article struct
		json.NewDecoder(rr.Body).Decode(&responseArticle)

		// Check the response string matches expected response
		if responseArticle.Id != 0 && responseArticle.Title != article.Title {
			t.Errorf("handler returned unexpected body: got %+v want %+v",
				responseArticle, article)
		}
	}
}

func TestReturnAllArticles(t *testing.T) {

	for _, dbType := range []string{"mysql", "postgres"} {

		var (
			connectionString string
			responseArticle  Article
		)

		if dbType == "mysql" {
			connectionString = MySQLConn
		} else {
			connectionString = PostgresConn
		}

		// start the DB connection and router for app
		app := initTestModule(
			dbType,
			connectionString,
		)

		// prepare article data
		article := Article{
			Title:   "TestArticle",
			Desc:    "TestDesc",
			Content: "TestContent",
		}

		// convert the article data as json
		payload, err := json.Marshal(article)
		if err != nil {
			t.Fatal(err)
		}

		// creating new request to homepage
		req, err := http.NewRequest("GET", "/articles", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatal(err)
		}

		// if db connection fails, add as fatal error
		if app.Database == nil || app.Database.Ping() != nil {
			t.Fatal(err)
		}

		// new recorder for capturing response from request
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(app.createNewArticle)

		// serve http call on request
		handler.ServeHTTP(rr, req)

		// checking http status code if 200
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		// decode the response body to a new article struct
		json.NewDecoder(rr.Body).Decode(&responseArticle)

		// Check the response string matches expected response
		if responseArticle.Id != 0 && responseArticle.Title != article.Title {
			t.Errorf("handler returned unexpected body: got %+v want %+v",
				responseArticle, article)
		}
	}
}
