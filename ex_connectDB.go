package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func query(db *sql.DB) {
	var (
		id      int
		detail  string
		income  float64
		expense float64
		date    string
	)

	var input_id int
	fmt.Scan(&input_id)

	query := "SELECT id, detail, income, expense, date FROM statement WHERE id = ?"
	if err := db.QueryRow(query, input_id).Scan(&id, &detail, &income, &expense, &date); err != nil {
		log.Fatal(err)
	}
	fmt.Println(id, detail, income, expense, date)
}

func Insert(db *sql.DB) {
	var detail string
	var income float64
	var expense float64
	fmt.Scan(&detail, &income, &expense)
	createAt := time.Now()

	result, err := db.Exec(`INSERT INTO statement (detail,income, expense, date) VALUES (?, ?, ?, ?)`, detail, income, expense, createAt)
	if err != nil {
		log.Fatal(err)
	}
	id, err := result.LastInsertId()
	fmt.Println(id)

}

func Delete(db *sql.DB)  {
	var delete_id int
	fmt.Scan(&delete_id)
	_, err := db.Exec(`DELETE FROM statement WHERE id = ?`, delete_id)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/myaccountdb")
	if err != nil {
		fmt.Println("failed to connect")
	} else {
		fmt.Println("connect successfully")
	}

	// query(db) //get data by id
	// Insert(db) //insert data
	Delete(db)
}
