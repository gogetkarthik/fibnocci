package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/urfave/cli"

	"github.com/fibonacci/pkg/fibonacci/flags"
)

//TODO convert this to flags
const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "example"
	dbname   = "postgres"
)

//TODO set this in context for each request.
//TODO define interface for DB interactions
var (
	db *sql.DB
)

type (
	maxUpdate bool
)

const (
	maxUpdateFalse maxUpdate = false
	maxUpdateTrue  maxUpdate = true
)

func NewFibonacci(appName string) *cli.App {
	return &cli.App{
		Name:        appName,
		Usage:       "app to find out fibonacci with memorization",
		Description: "app to find out fibonacci with memorization",
		Commands: []cli.Command{
			{
				Name:        "serve",
				ShortName:   "s",
				Usage:       "starts the fibonacci server",
				Description: "starts the fibonacci server",
				Before: func(ctx *cli.Context) error {
					psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
						"password=%s dbname=%s sslmode=disable",
						ctx.String("db-host"), port, user, password, dbname)
					var err error
					db, err = sql.Open("postgres", psqlInfo)
					if err != nil {
						panic(err)
					}

					if err = db.Ping(); err != nil {
						return err
					}

					return nil
				},
				After: func(context *cli.Context) error {
					defer db.Close()
					return nil
				},
				Action: func(ctx *cli.Context) error {
					http.HandleFunc("/favicon.ico", func(writer http.ResponseWriter, request *http.Request) {
						writer.WriteHeader(http.StatusOK)
					})

					//TODO use swagger to auto generate the server, client and model stubs
					http.HandleFunc("/", fib)
					if err := http.ListenAndServe(":9002", nil); err != nil {
						return err
					}
					return nil
				},
				Flags: []cli.Flag{
					flags.DBHost,
				},
			},
		},
		Compiled: time.Time{},
	}
}

func fib(w http.ResponseWriter, r *http.Request) {
	reg := "^\\/fib\\/\\d+$"
	match, err := regexp.Match(reg, []byte(r.URL.Path))
	if err != nil {
		apiResponseInternalServerError(w, err)
		return
	}

	if !match {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/fib/")
	rTxn, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  true,
	})
	if err != nil {
		apiResponseInternalServerError(w, err)
		return
	}

	maxFibKey, maxFib, err := getCurrentMaxFib(rTxn, maxUpdateFalse)
	if err != nil {
		apiResponseInternalServerError(w, err)
		return
	}

	if err := rTxn.Commit(); err != nil {
		apiResponseInternalServerError(w, err)
		return
	}

	fibToFind, err := strconv.Atoi(id)
	if err != nil {
		apiResponseInternalServerError(w, err)
		return
	}

	var value int
	if fibToFind > maxFibKey {
		//LevelReadCommitted is default behavior just calling explicitly for readability
		txn, err := db.BeginTx(context.Background(), &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
			ReadOnly:  false,
		})
		if err != nil {
			apiResponseInternalServerError(w, err)
			return
		}

		//TODO this call can be avoided using a caching layer
		maxFibKey, maxFib, err = getCurrentMaxFib(txn, maxUpdateTrue)
		if err != nil {
			apiResponseInternalServerError(w, err)
			return
		}

		if fibToFind > maxFibKey {
			mapFib := make(map[int]int)
			mapFib[maxFibKey] = maxFib
			maxFibLessOne := maxFibKey - 1
			maxFibLessValue, err := getFibValue(fmt.Sprintf("%d", maxFibLessOne))
			if err != nil {
				apiResponseInternalServerError(w, err)
			}
			mapFib[maxFibLessOne] = maxFibLessValue
			//Get it from DB to make it consistent with calculated and exiting values.
			_ = calCalculateFib(mapFib, fibToFind)

			delete(mapFib, maxFibKey)
			delete(mapFib, maxFibLessOne)
			err = createNewFibs(txn, mapFib)
			if err != nil {
				_ = txn.Rollback()
				apiResponseInternalServerError(w, err)
				return
			}

			err = updateMaxFib(txn, fibToFind, mapFib[fibToFind], maxFibKey)
			if err != nil {
				_ = txn.Rollback()
				apiResponseInternalServerError(w, err)
				return
			}

		}

		if txn != nil {
			if err := txn.Commit(); err != nil {
				apiResponseInternalServerError(w, err)
				return
			}
		}

	}

	value, err = getFibValue(id)
	if err != nil {
		apiResponseInternalServerError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "fibonacci for %d is %d", fibToFind, value)
}

func updateMaxFib(txn *sql.Tx, key, value, exitingMax int) error {
	var query = "update max set max_fib_key=%d, max_fib_value=%d where max_fib_key=%d"

	q := fmt.Sprintf(query, key, value, exitingMax)
	_ = q
	prepare, err := txn.Prepare(q)
	if err != nil {
		return err
	}

	results, err := prepare.Exec()
	if err != nil {
		return err
	}

	rowsAffected, err := results.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return errors.New("")
	}

	return nil
}

func createNewFibs(txn *sql.Tx, fibMap map[int]int) error {
	var query = "insert into fib(key, value) values %s"
	insertParam := `(%d, %d)`

	var bindInsertParam []string
	for key, val := range fibMap {
		bindInsertParam = append(bindInsertParam, fmt.Sprintf(insertParam, key, val))
	}

	q := fmt.Sprintf(query, strings.Join(bindInsertParam, ","))
	_ = q
	prepare, err := txn.Prepare(q)
	if err != nil {
		return err
	}

	results, err := prepare.Exec()
	if err != nil {
		return err
	}

	rowsAffected, err := results.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != int64(len(fibMap)) {
		return errors.New("")
	}

	return nil
}

func calCalculateFib(mapFib map[int]int, febToFindKey int) int {

	if val, ok := mapFib[febToFindKey]; ok {
		return val
	}

	mapFib[febToFindKey] = calCalculateFib(mapFib, febToFindKey-1) + calCalculateFib(mapFib, febToFindKey-2)

	return mapFib[febToFindKey]
}

func apiResponseInternalServerError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, err.Error())
}

func getFibValue(fibIndex string) (int, error) {
	row := db.QueryRow("select key, value from fib where key = " + fibIndex)

	var i, j *int

	err := row.Scan(&i, &j)
	if err != nil {
		return 0, err
	}
	return *j, nil
}

func getCurrentMaxFib(txn *sql.Tx, isMaxUpdate maxUpdate) (int, int, error) {

	query := "select max_fib_key, max_fib_value from max"

	if isMaxUpdate {
		query = fmt.Sprintf("%s for update", query)
	}

	row := txn.QueryRow(query)

	var maxFibKey, maxFibValue int
	row.Scan(&maxFibKey, &maxFibValue)

	return maxFibKey, maxFibValue, nil
}
