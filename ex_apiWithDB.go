package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var Db *sql.DB

const statementPath = "statements"
const summaryPath = "summary"
const basePath = "/api"

type Statement struct {
	Id      int     `json: "id"`
	Detail  string  `json: "detail"`
	Income  float64 `json: "income"`
	Expense float64 `json: "expense"`
	Date    string  `json: "date"`
}

type Summary struct {
	Sum_Income  float64 `json: "sum_income"`
	Sum_Expense float64 `json: "sum_expense"`
	Date        string  `json: "date"`
}

func SetupDb() {
	var err error
	Db, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/myaccountdb")

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("conncet database")
	Db.SetConnMaxLifetime(time.Minute * 30)
	Db.SetMaxOpenConns(10)
	Db.SetMaxIdleConns(10)
}

func getStatementList() ([]Statement, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := Db.QueryContext(ctx, ` SELECT id, detail, income, expense, date FROM statement`)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	defer result.Close()
	statements := make([]Statement, 0)
	for result.Next() {
		var statement Statement
		result.Scan(&statement.Id, &statement.Detail, &statement.Income, &statement.Expense, &statement.Date)
		statements = append(statements, statement)
	}
	return statements, nil
}

func insertStatement(statement Statement) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := Db.ExecContext(ctx, `INSERT INTO statement (detail,income, expense, date) VALUES (?, ?, ?, ?)`, statement.Detail, statement.Income, statement.Expense, statement.Date)
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}
	insertId, err := result.LastInsertId()
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}
	return int(insertId), nil
}

func getStatement(id int) (*Statement, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	row := Db.QueryRowContext(ctx, ` SELECT id, detail, income, expense, date FROM statement WHERE id = ?`, id)

	statement := &Statement{}
	err := row.Scan(&statement.Id, &statement.Detail, &statement.Income, &statement.Expense, &statement.Date)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	return statement, nil
}

func deleteStatement(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := Db.ExecContext(ctx, `DELETE FROM statement WHERE id = ?`, id)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func getDateSummary(date string) (*Summary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	row := Db.QueryRowContext(ctx, ` SELECT SUM(income) AS income, SUM(expense) AS expense,date FROM statement WHERE date = ?`, date)
	sum := &Summary{}
	err := row.Scan(&sum.Sum_Income, &sum.Sum_Expense, &sum.Date)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	return sum, nil
}

func handleStatements(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		statementList, err := getStatementList()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json, err := json.Marshal(statementList)
		if err != nil {
			log.Fatal(err)
		}
		_, err = w.Write(json)
		if err != nil {
			log.Fatal(err)
		}
	case http.MethodPost:
		var statement Statement
		err := json.NewDecoder(r.Body).Decode(&statement)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		StatementId, err := insertStatement(statement)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf(`{"statement_id: %d}`, StatementId)))
	case http.MethodOptions:
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleStatement(w http.ResponseWriter, r *http.Request) {
	urlPathSegments := strings.Split(r.URL.Path, fmt.Sprintf("%s/", statementPath))
	if len(urlPathSegments[1:]) > 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(urlPathSegments[len(urlPathSegments)-1])
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	switch r.Method {
	case http.MethodGet:
		statement, err := getStatement(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if statement == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json, err := json.Marshal(statement)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, err = w.Write(json)
		if err != nil {
			log.Fatal(err)
		}
	case http.MethodDelete:
		err := deleteStatement(id)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleSummary(w http.ResponseWriter, r *http.Request) {
	urlPathSegments := strings.Split(r.URL.Path, fmt.Sprintf("%s/", summaryPath))
	if len(urlPathSegments[1:]) > 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	date := urlPathSegments[len(urlPathSegments)-1]
	switch r.Method {
	case http.MethodGet:
		sum, err := getDateSummary(date)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if sum == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json, err := json.Marshal(sum)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, err = w.Write(json)
		if err != nil {
			log.Fatal(err)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func corsMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Add("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")
		handler.ServeHTTP(w, r)
	})
}

func SetupRoutes(apiBasePath string) {
	statementsHandler := http.HandlerFunc(handleStatements)
	statementHandler := http.HandlerFunc(handleStatement)
	summaryHandler := http.HandlerFunc(handleSummary)

	http.Handle(fmt.Sprintf("%s/%s", apiBasePath, statementPath), corsMiddleware(statementsHandler))
	http.Handle(fmt.Sprintf("%s/%s/", apiBasePath, statementPath), corsMiddleware(statementHandler))
	http.Handle(fmt.Sprintf("%s/%s/", apiBasePath, summaryPath), corsMiddleware(summaryHandler))
}

func main() {
	SetupDb()
	SetupRoutes(basePath)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
