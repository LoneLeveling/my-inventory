package main

import (
	"database/sql"
	"errors"
	"fmt"
)

type product struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

func getProducts(db *sql.DB) ([]product, error) {
	//fetching all DB rows below:
	query := "SELECT id, name, quantity, price from products"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	//If no error then create an empty slice products, and loop on each row to fetch the data from the DB table 'products'
	products := []product{}
	for rows.Next() {
		var p product
		err := rows.Scan(&p.ID, &p.Name, &p.Quantity, &p.Price)
		if err != nil {
			return nil, err
		}
		products = append(products, p) //appending the object to the slice if no error found
	}
	return products, nil //returning the entire slice and err as nil, since no error found.

}

// IMP NOTE:In Go, %v only works when you pass the value through fmt.Sprintf
func (p *product) getProduct(db *sql.DB) error {
	query := fmt.Sprintf("SELECT name, quantity,price FROM products where id=%v", p.ID)
	//Using query row method from db pointer to get the complete row data
	//NOTE: QueryRow() is used when there is atmost one row that needs to be returned.
	row := db.QueryRow(query)
	err := row.Scan(&p.Name, &p.Quantity, &p.Price)
	if err != nil {
		return err
	}
	return nil
}

// We receive data below from Route Handler in app.go
func (p *product) createProduct(db *sql.DB) error {
	query := fmt.Sprintf("insert into products(name, quantity, price) values ('%v',%v,%v)", p.Name, p.Quantity, p.Price) //NOTE: Name is string value so we kept %v in quotes
	result, err := db.Exec(query)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	p.ID = int(id)
	return nil

}

// Below func executes a db query in the database
func (p *product) updateProduct(db *sql.DB) error {
	query := fmt.Sprintf("Update products set name='%v', quantity=%v, price=%v where id=%v", p.Name, p.Quantity, p.Price, p.ID) //NOTE: Name is string value so we kept %v in quotes
	result, err := db.Exec(query)
	//Exec() executes the above SQL query in the database
	rowsAffected, err := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("No such row exists")
	}
	return err
}

func (p *product) deleteProduct(db *sql.DB) error {
	query := fmt.Sprintf("Delete from products where id=%v", p.ID)
	_, err := db.Exec(query)
	return err
}
