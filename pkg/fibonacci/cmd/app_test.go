package cmd

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func Test_fib_pathNotImplemented(t *testing.T) {
	testDB, _, err := sqlmock.New()
	if err != nil {
		assert.Nil(t, err)
	}
	defer testDB.Close()
	db = testDB
	req := httptest.NewRequest(http.MethodPost, "http://localhost:9001/fib/", nil)
	rr := httptest.NewRecorder()

	fib(rr, req)
	assert.Equal(t, http.StatusNotImplemented, rr.Result().StatusCode)
}

func Test_fib_beginTxnFailed(t *testing.T) {
	testDB, mock, err := sqlmock.New()
	if err != nil {
		assert.Nil(t, err)
	}
	defer testDB.Close()
	db = testDB
	db.Begin()
	req := httptest.NewRequest(http.MethodPost, "http://localhost:9001/fib/10", nil)
	rr := httptest.NewRecorder()
	mock.ExpectBegin().WillReturnError(errors.New("begin tx failed"))

	fib(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Result().StatusCode)
}

