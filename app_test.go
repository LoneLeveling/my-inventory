package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

var a App

func TestMain(m *testing.M) {

	err := a.Initialise(DBUser, DbPassword, "test")
	if err != nil {
		log.Fatal("Error occured while intialising the database")
	}
	createTable()
	m.Run()
}

func createTable() {
	createTableQuery := `CREATE TABLE IF NOT EXISTS products (
		id int NOT NULL AUTO_INCREMENT,
		name varchar(255) NOT NULL,
		quantity int,
		price float(10,7),
		PRIMARY KEY (id)
	);`
	_, err := a.DB.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	a.DB.Exec("DELETE FROM products")
	a.DB.Exec("ALTER TABLE products AUTO_INCREMENT=1")
}

func addProduct(name string, quantity int, price float64) {
	query := fmt.Sprintf("INSERT INTO products(name, quantity, price) VALUES('%v', %v, %v)", name, quantity, price)
	_, err := a.DB.Exec(query)
	if err != nil {
		log.Printf("Error adding product: %v", err)
	}
}
func TestGetProduct(t *testing.T) {
	clearTable()
	addProduct("chair", 1, 800.00)
	request, _ := http.NewRequest("GET", "/product/1", nil)
	response := sendRequest(request)
	checkStatusCode(t, http.StatusOK, response.Code)
}

func checkStatusCode(t *testing.T, expectedStatusCode int, actualStatusCode int) {
	if expectedStatusCode != actualStatusCode {
		t.Errorf("Expected status: %v, Received: %v", expectedStatusCode, actualStatusCode)
	}
}

func sendRequest(request *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	a.Router.ServeHTTP(recorder, request)
	return recorder
}

func TestCreateProduct(t *testing.T) {
	clearTable()
	var product = []byte(`{"name":"chair", "quantity":1, "price":100}`)
	req, _ := http.NewRequest("POST", "/product/", bytes.NewBuffer(product))
	req.Header.Set("Content-Type", "application/json")

	response := sendRequest(req)
	checkStatusCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "chair" {
		t.Errorf("Expected name: %v, Got: %v", "chair", m["name"])
	}
	// log.Printf("%T", m["quantity"])
	if m["quantity"] != 1.0 { // Numbers are unmarshaled as float64.
		t.Errorf("Expected quantity: %v, Got: %v", 1.0, m["quantity"])
	}
}

func TestDeleteProduct(t *testing.T) {
	//1.We clear everything from the table 1st
	//2.We going to add data into the table
	//3.We will check if we are able to fetch the added products
	//4.Then we execute the DELETE operation on product added
	//5. Lastly try to fetch data and see if delete opeation worked fine.

	clearTable()
	addProduct("HDMI cable", 4, 950)

	//GET call to fetch the product
	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := sendRequest(req)
	checkStatusCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/product/1", nil)
	response = sendRequest(req)
	checkStatusCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/product/1", nil)
	response = sendRequest(req)
	checkStatusCode(t, http.StatusNotFound, response.Code)
	//NOTE: In above line if we use:http.StatusOk instead, then Running: go test , gives :Expected status: 200, Received: 404 , **RMBR: We received 404 because resource could not be found as we deleted everything from the table.
}

// Testing out PUT http verb
func TestUpdateProduct(t *testing.T) {
	// 1. 1st we clear everything from the table and then add data into it.
	// 2. Use GET API to fetch that product
	// 3. Update that product using PUT endpoint.
	//4. Save the response received from step 4 and compare the two responsed against each other.

	clearTable()
	addProduct("HDMI cable", 4, 950)
	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := sendRequest(req)
	checkStatusCode(t, http.StatusOK, response.Code)
	//We now need to save the response post adding data into table for later comparisons that we need to do.
	//So everything will be stored in this 'oldValue' map below
	var oldValue map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &oldValue)

	//Now using PUT end point lets change values around our product
	var product = []byte(`{"name":"HDMI cable", "quantity":2, "price":475}`)
	req, _ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(product))
	req.Header.Set("Content-Type", "application/json")

	response = sendRequest(req)

	//Storing new value
	var newValue map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &newValue)

	//Comparing oldValue(complete Row) Vs newValue(complete Row)
	if oldValue["id"] != newValue["id"] {
		t.Errorf("Expected id:%v,Got:%v", newValue["id"], oldValue["id"])
	}

	//Testing for name
	if oldValue["name"] != newValue["name"] {
		t.Errorf("Expected id:%v,Got:%v", newValue["name"], oldValue["name"])
	}

	//Testing for Price,Throwing error if price reamins same
	if oldValue["price"] == newValue["price"] {
		t.Errorf("Expected id:%v,Got:%v", newValue["price"], oldValue["price"])
	}

	//Testing for Quantity, Throwing error if quantities reamins same
	if oldValue["quantity"] == newValue["quantity"] {
		t.Errorf("Expected id:%v,Got:%v", newValue["quantity"], oldValue["quantity"])
	}
}
