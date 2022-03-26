package handler

import (
	"database/sql"
	"discordgo-blackjack/data"
	"fmt"
)

// BaseHandler Holds everything controller needs
type BaseHandler struct {
	db *sql.DB
}

// NewBaseHandler Returns a new instance of a base handler
func NewBaseHandler(db *sql.DB) *BaseHandler {
	return &BaseHandler{
		db: db,
	}
}

func (handler *BaseHandler) CheckPing() {
	if err := handler.db.Ping(); err != nil {
		fmt.Println("Error pinging database")
	}
}

func (handler *BaseHandler) GetDBConn() *sql.DB {
	return handler.db
}

func (handler *BaseHandler) CloseDBConn() {
	handler.db.Close()
}

func (handler *BaseHandler) OpenDBConn() {
	handler.db, _ = sql.Open("postgres", data.GetDBConnString())
	if err := handler.db.Ping(); err != nil {
		fmt.Println("Error opening DB Connection")
	}
}
