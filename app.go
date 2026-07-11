// app.go file stores information about our app variables such as router and DB variables, and also some methods related to handling our routes.
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// Below struct stores 2 infos. Router info & DB info.
type App struct {
	Router *mux.Router
	DB     *sql.DB
}

//Creating 2 methods for struct:

func (app *App) Initialise(DBUser string, DbPassword string, DBName string) error {
	//1. Opening DB connection
	connectionString := fmt.Sprintf("%v:%v@tcp(127.0.0.1:3306)/%v", DBUser, DbPassword, DBName) //Produces: root:qawsed@123@tcp(127.0.0.1:3306)/inventory

	var err error
	app.DB, err = sql.Open("mysql", connectionString)
	if err != nil {
		return err
	}

	// 2. Creating our HTTP Router
	app.Router = mux.NewRouter().StrictSlash(true)
	//calling handleRoutes() after creating a router DB connection
	app.handleRoutes()
	return nil
}

// Second method is also created on our app struct Pointer
func (app *App) Run(address string) {
	log.Fatal(http.ListenAndServe(address, app.Router))
}

func sendResponse(w http.ResponseWriter, statusCode int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
}

func SendError(w http.ResponseWriter, statusCode int, err string) {
	error_message := map[string]string{"error": err}
	sendResponse(w, statusCode, error_message)
}

func (app *App) getProducts(w http.ResponseWriter, r *http.Request) {
	products, err := getProducts(app.DB) //This calls model.go's(Db interacting file) getProducts()
	if err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendResponse(w, http.StatusOK, products)
}

func (app *App) getProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key, err := strconv.Atoi(vars["id"])
	if err != nil {
		SendError(w, http.StatusBadRequest, "invalid product ID")
		return
	}

	p := product{ID: key} //product struct
	err = p.getProduct(app.DB)
	if err != nil {
		//NOTE: 2 error types possibility here, No.1: row doesnt exit in DB, 2nd: Internal server error, so we use a switch statement to handle that.
		switch err {
		case sql.ErrNoRows:
			SendError(w, http.StatusNotFound, "Product not found")
		default:
			SendError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	sendResponse(w, http.StatusOK, p)

}

func (app *App) createProduct(w http.ResponseWriter, r *http.Request) {
	//Now in POST request client send data which is in json format, to be able to store that in DB we need to decode it, as shown below:
	var p product //struct variable p
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		SendError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err = p.createProduct(app.DB)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	//And if all went well, below line gets executed:
	sendResponse(w, http.StatusCreated, p)
}

func (app *App) updateProduct(w http.ResponseWriter, r *http.Request) {
	// 1. First we understand for which {id} user want to update the details for.
	//So converting the id received into integer using Atoi(ASCII to Integer) method from strconv package.
	vars := mux.Vars(r)
	key, err := strconv.Atoi(vars["id"])
	if err != nil {
		SendError(w, http.StatusBadRequest, "invalid product ID")
		return
	}

	//2. Taking the json input received from the user and converting it into structured variable.
	var p product //struct variable p
	err = json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		SendError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	p.ID = key
	err = p.updateProduct(app.DB)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendResponse(w, http.StatusOK, p)

}

func (app *App) deleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key, err := strconv.Atoi(vars["id"])
	if err != nil {
		SendError(w, http.StatusBadRequest, "invalid product ID")
		return
	}
	p := product{ID: key}
	err = p.deleteProduct(app.DB)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendResponse(w, http.StatusOK, map[string]string{
		"result": fmt.Sprintf("succesfully deleted row with ID=%v", key),
	})
}

// Method to handle all our Routes
func (app *App) handleRoutes() {
	app.Router.HandleFunc("/products", app.getProducts).Methods("GET")          //fetching all products
	app.Router.HandleFunc("/product/{id}", app.getProduct).Methods("GET")       //fetching a product via id if it exists
	app.Router.HandleFunc("/product/", app.createProduct).Methods("POST")       //Adding a product into database
	app.Router.HandleFunc("/product/{id}", app.updateProduct).Methods("PUT")    //PUT: Used to update an already existing product in database.
	app.Router.HandleFunc("/product/{id}", app.deleteProduct).Methods("DELETE") //DELETE: Used to delete an already existing product in database.
}
