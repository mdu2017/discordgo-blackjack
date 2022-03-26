package data

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Database connection
var DBConn *sql.DB

// Opens a database connection
func OpenDBConnection() *sql.DB {

	dbConnString := GetDBConnString()

	db, err := sql.Open("postgres", dbConnString)
	if err != nil {
		fmt.Println(dbConnString)
		panic(err)
	}

	// Leave this off since the hosting platform will take care of it
	//defer DBConn.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to database")

	return db
}

func GetDBConnString() string {

	// ============ Database credentials ==============
	fp, err := os.Open("data/conn.txt")
	defer fp.Close()
	if err != nil {
		panic("can't open credentials file")
	}

	scanner := bufio.NewScanner(fp)
	scanner.Scan() // Scan first line into buffer
	credentialString := scanner.Text()

	dbCredentials := strings.Split(credentialString, ",")

	// NOTE: When hosting, add the "sslmode=disable", when running locally, allow ssl
	portNum, _ := strconv.Atoi(dbCredentials[1])

	return fmt.Sprintf("host=%s port=%d user=%s password=%s database=%s",
		dbCredentials[0], portNum, dbCredentials[2], dbCredentials[3], dbCredentials[4])
}
