package token

import (
	"errors"
	"fmt"
	"os"
	"time"

	"loltracking-api/helper"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type TokenRow struct {
	id       int
	token    string
	isActive int
}

var (
	db    *sqlx.DB
	blank interface{}
)

func CheckToken(token string) (errResp interface{}) {
	if len(token) > 0 {
		if db, errResp = setAndGetDb(); errResp != blank {
			return errResp
		}

		queryString := "SELECT id, token, is_active FROM settings WHERE token = ? LIMIT 1"
		row := db.QueryRow(queryString, token)

		var tokenRow TokenRow

		err := row.Scan(&tokenRow.id, &tokenRow.token, &tokenRow.isActive)
		if errResp = helper.ErrorResponse(err); errResp != blank {
			return tokenResponse("Sorry, not allowed.")
		}

		if tokenRow.isActive == 0 {
			return tokenResponse("Token is expired.")
		}

		/*if errResp := setTokenExpired(tokenRow.id); errResp != blank {
			return errResp
		}*/

		return blank
	}

	return tokenResponse("Sorry, not allowed.")
}

func setTokenExpired(tokenId int) interface{} {
	date := time.Now()
	dateFormat := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute(), date.Second())
	queryString := "UPDATE settings SET is_active = ?, date_expired = ? WHERE id = ?"
	result, err := db.Exec(queryString, 0, dateFormat, tokenId)
	if errResp := helper.ErrorResponse(err); errResp != blank {
		return errResp
	}

	if affectedRows, err := result.RowsAffected(); affectedRows > 0 {
		if errResp := helper.ErrorResponse(err); errResp != blank {
			return errResp
		}
		return blank
	}

	return helper.ErrorResponse(errors.New("Bir problem meydana geldi."))
}

func tokenResponse(message string) interface{} {
	fatalResponse := helper.SetAndGetResponse(false, message, nil, helper.HttpStatusForbidden)
	return fatalResponse
}

func setAndGetDb() (db *sqlx.DB, errResp interface{}) {
	var err error
	db, err = sqlx.Connect("mysql", getDbCredentials())
	if errResp := helper.ErrorResponse(err); errResp != blank {
		return nil, errResp
	}

	return db, errResp
}

func getDbCredentials() string {
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")

	//println(dbUser + ":" + dbPass + "@tcp(" + dbHost + ")/" + dbName)
	//db_user:password@tcp(localhost:3306)/my_db
	return dbUser + ":" + dbPass + "@tcp(" + dbHost + ")/" + dbName
}
