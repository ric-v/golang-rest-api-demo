/*
REST API DEMO

	createNewArticle
	returnAllArticles
	returnSingleArticle
	updateArticle
	homepage
	deleteArticle
	handleRequests
	connectToDB
	main
*/
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type (

	// App controlls the rest API demo app
	App struct {
		Router   *mux.Router
		Database *sql.DB
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

	var article Article

	log.Println("Endpoint hit : createNewArticle")
	// get the payload from request
	err := json.NewDecoder(r.Body).Decode(&article)
	if err != nil {
		fmt.Println(err)
	}

	// insert data into DB
	response, err := app.Database.Exec(
		"INSERT INTO articles (title, descr, content) VALUES (?,?,?)",
		article.Title, article.Desc, article.Content,
	)
	// if there is an error inserting, handle it
	if err != nil {
		panic(err.Error())
	}
	log.Print(response.RowsAffected())
	log.Println("inserted new record to DB")

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

	log.Println("Endpoint hit : returnAllArticles")

	// get the id and limit from param
	lastID := r.URL.Query().Get("id")
	limit := r.URL.Query().Get("limit")

	// if last id is empty, set as 0
	if lastID == "" {
		lastID = "0"
	}

	// if limti is empty, get all entries else get all entries with limit
	if limit == "" {
		query = "SELECT * FROM articles WHERE id > ? ORDER BY id ASC"
		queryParams = append(queryParams, lastID)
	} else {
		query = "SELECT * FROM articles WHERE id > ? ORDER BY id ASC LIMIT ?"
		queryParams = append(queryParams, lastID, limit)
	}
	log.Println(query, queryParams)

	// insert data into DB
	response, err := app.Database.Query(
		query,
		queryParams...,
	)
	// if there is an error inserting, handle it
	if err != nil {
		panic(err.Error())
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
			panic(err.Error())
		}

		// append to final list of articles
		articles = append(articles, article)
	}
	log.Printf("article : %+v\n", articles)

	// generate JSON resopnse
	err = json.NewEncoder(w).Encode(articles)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Endpoint hit : return all articles")
}

//	GET /returnSingleArticle/{id}
//	url params : id (article ID to be retrieved)
//	response   : Article struct
//
// return a selected article value from DB
func (app *App) returnSingleArticle(w http.ResponseWriter, r *http.Request) {

	var article Article

	log.Println("Endpoint hit : returnSingleArticle")
	// get url path parameters
	vars := mux.Vars(r)
	key := vars["id"]

	// insert data into DB
	response, err := app.Database.Query(
		"SELECT * FROM articles WHERE id=?",
		key,
	)
	// if there is an error inserting, handle it
	if err != nil {
		panic(err.Error())
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
			panic(err.Error())
		}
	}
	log.Printf("article : %+v\n", article)

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

	log.Println("Endpoint hit : updateArticle")
	// get the path parameter
	vars := mux.Vars(r)
	key := vars["id"]

	var updatedArticle Article

	// get the payload data for article
	err := json.NewDecoder(r.Body).Decode(&updatedArticle)
	if err != nil {
		fmt.Println(err)
	}

	// update data in DB
	response, err := app.Database.Exec(
		"UPDATE articles SET title=?, descr=?, content=? WHERE id=?",
		updatedArticle.Title,
		updatedArticle.Desc,
		updatedArticle.Content,
		key,
	)
	// if there is an error inserting, handle it
	if err != nil {
		panic(err.Error())
	}
	log.Print(response.RowsAffected())
	log.Println(" mysql update performed.")

	// return the JSON response for added article
	json.NewEncoder(w).Encode(updatedArticle)
}

//	DELETE /deleteArticle/{id}
//	url params : id (article ID to be retrieved)
//
// remove an article from DB
func (app *App) deleteArticle(w http.ResponseWriter, r *http.Request) {

	log.Println("Endpoint hit : deleteArticle")
	// get url path parameter
	vars := mux.Vars(r)
	key := vars["id"]

	// insert data into DB
	response, err := app.Database.Exec(
		"DELETE FROM articles WHERE id=?",
		key,
	)
	// if there is an error inserting, handle it
	if err != nil {
		panic(err.Error())
	}
	log.Print(response.RowsAffected())
	log.Println(" mysql delete performed.")
}

//	ANY /homepage
//
// home page of web server
func (app *App) homepage(w http.ResponseWriter, r *http.Request) {

	log.Println("Endpoint hit : homepage")
	fmt.Fprint(w, "Welcome to the homepage!")
}

// http handler methods init
func handleRequests(app *App) {

	// start the gorilla mux router
	app.Router = mux.NewRouter().StrictSlash(true)

	// http routes
	app.Router.HandleFunc("/", app.homepage)
	app.Router.HandleFunc("/articles", app.returnAllArticles)
	app.Router.HandleFunc("/article", app.createNewArticle).Methods("POST")
	app.Router.HandleFunc("/article/{id}", app.updateArticle).Methods("PUT")
	app.Router.HandleFunc("/article/{id}", app.deleteArticle).Methods("DELETE")
	app.Router.HandleFunc("/article/{id}", app.returnSingleArticle).Methods("GET")

	// start the server on port
	log.Fatal(http.ListenAndServe(":7777", app.Router))
}

// establish DB connection for mysql DB
func connectToDB() (db *sql.DB, err error) {

	const connectionString = "myuser:mypass@tcp(localhost:3306)/db"

	// establish new mysql db connection
	db, err = sql.Open("mysql", connectionString)

	// if there is an error opening the connection, handle it
	if err != nil {
		panic(err.Error())
	}

	// execute a ping on mysql DB
	err = db.Ping()

	// if there is an error opening the connection, handle it
	if err != nil {
		panic(err.Error())
	}
	log.Println("Established mysql DB connection for ", connectionString)
	return
}

// main function
func main() {

	// connect to DB
	dbConn, err := connectToDB()
	if err != nil {
		fmt.Println(err)
	}

	// set new router
	app := &App{
		Router:   mux.NewRouter().StrictSlash(true),
		Database: dbConn,
	}

	// defer the close till after the main function has finished
	// executing
	defer app.Database.Close()

	// initialize the routes for rest API server
	handleRequests(app)
}
