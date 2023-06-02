package main

import (
	"database/sql"
	"fmt"
	"log"

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

func main() {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/myaccountdb")
	if err != nil {
		fmt.Println("failed to connect")
	} else {
		fmt.Println("connect successfully")
	}
	// fmt.Println(db)
	query(db)
}
