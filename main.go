/*
REST API DEMO

Methods :
createNewArticle, returnAllArticles, returnSingleArticle, updateArticle, homepage, deleteArticle, handleRequests, connectToDB, main
*/
package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type (

	// App controlls the rest API demo app
	App struct {
		DBType   string
		Router   *mux.Router
		Database *sql.DB
		logger   *log.Logger
	}

	// Article contains the data to be details for data to be stored into DB
	Article struct {
		Id      int    `json:"id"`
		Title   string `json:"Title"`
		Desc    string `json:"desc"`
		Content string `json:"content"`
	}
)

//	POST /createNewArticle
//	payload : Article struct
//
// creates new article entry to DB
func (app *App) createNewArticle(w http.ResponseWriter, r *http.Request) {

	var (
		query   string
		article Article
	)

	app.logger.Println("Endpoint hit : createNewArticle")
	// get the payload from request
	err := json.NewDecoder(r.Body).Decode(&article)
	if err != nil {
		app.logger.Println(err)
	}

	if app.DBType == "mysql" {
		query = "INSERT INTO articles (title, descr, content) VALUES (?,?,?)"
	} else if app.DBType == "postgres" {
		query = "INSERT INTO articles (title, descr, content) VALUES ($1,$2,$3)"
	}

	// insert data into DB
	response, err := app.Database.Exec(query, article.Title, article.Desc, article.Content)
	// if there is an error inserting, handle it
	if err != nil {
		app.logger.Println(err.Error())
		return
	}
	app.logger.Print(response.RowsAffected())
	app.logger.Println("inserted new record to DB")

	// return the added article
	json.NewEncoder(w).Encode(article)
}

//	GET /returnAllArticles
//	query params : id (last displayed ID for pagination), limit (max entry count in display)
//	response     : Article struct array
//
// get all the articles from DB
func (app *App) returnAllArticles(w http.ResponseWriter, r *http.Request) {

	var (
		query       string
		queryParams []interface{}
		articles    []Article
	)

	app.logger.Println("Endpoint hit : returnAllArticles")

	// get the id and limit from param
	lastID := r.URL.Query().Get("id")
	limit := r.URL.Query().Get("limit")

	// if last id is empty, set as 0
	if lastID == "" {
		lastID = "0"
	}

	// if limti is empty, get all entries else get all entries with limit
	if limit == "" {

		if app.DBType == "mysql" {
			query = "SELECT * FROM articles WHERE id > ? ORDER BY id ASC"
		} else if app.DBType == "postgres" {
			query = "SELECT * FROM articles WHERE id > $1 ORDER BY id ASC"
		}
		queryParams = append(queryParams, lastID)
	} else {

		if app.DBType == "mysql" {
			query = "SELECT * FROM articles WHERE id > ? ORDER BY id ASC LIMIT ?"
		} else if app.DBType == "postgres" {
			query = "SELECT * FROM articles WHERE id > $1 ORDER BY id ASC LIMIT $2"
		}
		queryParams = append(queryParams, lastID, limit)
	}
	app.logger.Println(query, queryParams)

	// insert data into DB
	response, err := app.Database.Query(
		query,
		queryParams...,
	)
	// if there is an error inserting, handle it
	if err != nil {
		app.logger.Println(err.Error())
		return
	}
	defer response.Close()

	// get all records until all are read
	for response.Next() {

		var article Article

		// get data from DB for article fields
		err = response.Scan(
			&article.Id,
			&article.Title,
			&article.Desc,
			&article.Content,
		)
		// if there is an error inserting, handle it
		if err != nil {
			app.logger.Println(err.Error())
			return
		}

		// append to final list of articles
		articles = append(articles, article)
	}
	app.logger.Printf("article : %+v\n", articles)

	// generate JSON resopnse
	err = json.NewEncoder(w).Encode(articles)
	if err != nil {
		app.logger.Println(err)
	}
	app.logger.Println("Endpoint hit : return all articles")
}

//	GET /returnSingleArticle/{id}
//	url params : id (article ID to be retrieved)
//	response   : Article struct
//
// return a selected article value from DB
func (app *App) returnSingleArticle(w http.ResponseWriter, r *http.Request) {

	var (
		query   string
		article Article
	)

	app.logger.Println("Endpoint hit : returnSingleArticle")
	// get url path parameters
	vars := mux.Vars(r)
	key := vars["id"]

	if app.DBType == "mysql" {
		query = "SELECT * FROM articles WHERE id=?"
	} else if app.DBType == "postgres" {
		query = "SELECT * FROM articles WHERE id=$1"
	}

	// insert data into DB
	response, err := app.Database.Query(query, key)
	// if there is an error inserting, handle it
	if err != nil {
		app.logger.Println(err.Error())
		return
	}
	defer response.Close()

	// iterate until entries from db are read
	for response.Next() {

		// scan and get article fields value
		err = response.Scan(
			&article.Id,
			&article.Title,
			&article.Desc,
			&article.Content,
		)
		// if there is an error inserting, handle it
		if err != nil {
			app.logger.Println(err.Error())
			return
		}
	}
	app.logger.Printf("article : %+v\n", article)

	// if article ID is not empty, return JSON response
	if article.Id != 0 {
		json.NewEncoder(w).Encode(article)
	} else {
		http.Error(w, "no record", http.StatusNotFound)
	}
}

//	PUT /updateArticle/{id}
//	url params : id (article ID to be retrieved)
//
// update the article for a given article ID
func (app *App) updateArticle(w http.ResponseWriter, r *http.Request) {

	var (
		query          string
		updatedArticle Article
	)

	app.logger.Println("Endpoint hit : updateArticle")
	// get the path parameter
	vars := mux.Vars(r)
	key := vars["id"]

	// get the payload data for article
	err := json.NewDecoder(r.Body).Decode(&updatedArticle)
	if err != nil {
		app.logger.Println(err)
	}

	if app.DBType == "mysql" {
		query = "UPDATE articles SET title=?, descr=?, content=? WHERE id=?"
	} else if app.DBType == "postgres" {
		query = "UPDATE articles SET title=$1, descr=$2, content=$3 WHERE id=$4"
	}

	// update data in DB
	response, err := app.Database.Exec(query, updatedArticle.Title, updatedArticle.Desc, updatedArticle.Content, key)
	// if there is an error inserting, handle it
	if err != nil {
		app.logger.Println(err.Error())
		return
	}
	app.logger.Print(response.RowsAffected())
	app.logger.Println(" DB update performed.")

	// return the JSON response for added article
	json.NewEncoder(w).Encode(updatedArticle)
}

//	DELETE /deleteArticle/{id}
//	url params : id (article ID to be retrieved)
//
// remove an article from DB
func (app *App) deleteArticle(w http.ResponseWriter, r *http.Request) {

	var query string

	app.logger.Println("Endpoint hit : deleteArticle")
	// get url path parameter
	vars := mux.Vars(r)
	key := vars["id"]

	if app.DBType == "mysql" {
		query = "DELETE FROM articles WHERE id=?"
	} else if app.DBType == "postgres" {
		query = "DELETE FROM articles WHERE id=$1"
	}

	// insert data into DB
	response, err := app.Database.Exec(query, key)
	// if there is an error inserting, handle it
	if err != nil {
		app.logger.Println(err.Error())
		return
	}
	app.logger.Print(response.RowsAffected())
	app.logger.Println(" DB delete performed.")
}

//	ANY /homepage
//
// home page of web server
func (app *App) homepage(w http.ResponseWriter, r *http.Request) {

	app.logger.Println("Endpoint hit : homepage")
	fmt.Fprint(w, `
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
`)
}

// http handler methods init
func handleRequests(app *App, port string) {

	// start the gorilla mux router
	app.Router = mux.NewRouter().StrictSlash(true)

	// http routes
	app.Router.HandleFunc("/", app.homepage)
	app.Router.HandleFunc("/articles", app.returnAllArticles).Methods("GET")
	app.Router.HandleFunc("/article", app.createNewArticle).Methods("POST")
	app.Router.HandleFunc("/article/{id}", app.updateArticle).Methods("PUT")
	app.Router.HandleFunc("/article/{id}", app.deleteArticle).Methods("DELETE")
	app.Router.HandleFunc("/article/{id}", app.returnSingleArticle).Methods("GET")

	// start the server on port
	app.logger.Fatal(http.ListenAndServe(":"+port, app.Router))
}

// establish DB connection for mysql DB
func connectToDB(dbType, connectionString string, logger *log.Logger) (db *sql.DB, err error) {

	// establish new db connection
	db, err = sql.Open(dbType, connectionString)

	// if there is an error opening the connection, handle it
	if err != nil {
		logger.Println(err.Error())
		return
	}

	// execute a ping on DB
	err = db.Ping()

	// if there is an error opening the connection, handle it
	if err != nil {
		logger.Println(err.Error())
		return
	}
	logger.Println("Established "+dbType+" DB connection for ", connectionString)
	return
}

// main function
func main() {

	var connectionString string

	dbType := flag.String("dbtype", "mysql", "choose the database type : mysql / postgres")
	dbUser := flag.String("user", "root", "db user name")
	dbPass := flag.String("pass", "my-secret-pw", "db password")
	dbHost := flag.String("host", "localhost", "host address for DB")
	dbPort := flag.String("dbport", "3306", "port of the running DB")
	dbName := flag.String("db", "db", "database name for connection")
	port := flag.String("port", "7777", "port for running this service")
	flag.Parse()

	// based on the db type set the connection string
	if *dbType == "mysql" {

		connectionString = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", *dbUser, *dbPass, *dbHost, *dbPort, *dbName)

	} else if *dbType == "postgres" {

		connectionString = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			*dbHost, *dbPort, *dbUser, *dbPass, *dbName,
		)
	}

	// store the log file data to log file
	logFile, _ := os.OpenFile(
		"./restful_api.log",
		os.O_TRUNC|os.O_CREATE|os.O_RDWR,
		os.ModePerm,
	)

	logger := log.New(
		logFile,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile,
	)

	// connect to DB
	dbConn, err := connectToDB(*dbType, connectionString, logger)
	if err != nil {
		log.Println(err)
	}

	// set new router
	app := &App{
		DBType:   *dbType,
		Router:   mux.NewRouter().StrictSlash(true),
		Database: dbConn,
		logger:   logger,
	}

	// defer the close till after the main function has finished
	// executing
	defer app.Database.Close()

	// initialize the routes for rest API server
	handleRequests(app, *port)
}
