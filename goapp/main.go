package main

import (
	"database/sql"
	"goapp/sqlite"
	"goapp/transactions"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

const crdbMode = "CRDB"
const sqliteMode = "SQLITE"
const mockedAccountID = "90903a90-d8f0-45eb-a4aa-dea4d24b2f54"

func main() {
	var db *sql.DB
	var err error
	var parallelize = false

	mode := os.Getenv("MODE")
	log.Printf("Starting in %q mode", mode)

	switch mode {
	case crdbMode:
		// Connect to one of the three nodes by passing through an HAProxy Load Balancer.
		db, err = sql.Open("postgres", "postgresql://myuser@localhost:26257/bank?sslmode=disable")
	case sqliteMode:
		parallelize = false
		db, err = sqlite.New("sqlite.db?cache=private&_foreign_keys=1&_txlock=exclusive&_journal=DELETE")
	default:
		log.Fatalf("Mode %q out of the known domain", mode)
	}

	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Could not close db: %v", err)
		}
	}()

	/**
	 * 1 = 27.16s / 32.59s / 24.53s / 26.76s / 29.79s = ~28.17s
	 * 10 = 31.48s / 27.96s / 32.04s / 28.15s / 27.62s = ~29.45s
	 * 100 = 34.08s / 35.01s / 30.94s / 35.07s / 32.40s = ~33.50s
	 * 1000 = 26.43s / 38.44s / 29.64s / 27.38s / 30.85s = ~30.55s
	 */
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Setting up Monetary transactions that we want to execute within the same Database transaction
	list := []transactions.Transaction{
		{
			AmountCents: 50000,
			Description: "First transaction",
		},
		{
			AmountCents: 49500,
			Description: "Second transaction",
		},
	}

	delta := 2000
	start := time.Now()

	wg := sync.WaitGroup{}
	wg.Add(delta)
	for i := 0; i < delta; i++ {
		// delta-1 out of delta transactions are going to fail
		go func() {
			defer wg.Done()
			switch mode {
			case crdbMode:
				if err := transactions.ProcessList(db, mockedAccountID, list, parallelize); err != nil {
					log.Printf("Could not process transactions list: %v", err)
				}
			case sqliteMode:
				if err := transactions.ProcessListSqlite(db, mockedAccountID, list); err != nil {
					log.Printf("Could not process transactions list: %v", err)
				}
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)
	log.Printf("Script took %s\n", elapsed)
	log.Printf("DB Stats: %+v\n", db.Stats())

	// Print out the balance, it should show 5€ left (i.e. 500).
	row := db.QueryRow("SELECT balance_cents FROM accounts WHERE id = $1", mockedAccountID)

	var balance int64
	if err := row.Scan(&balance); err != nil {
		log.Fatalf("Could not scan row: %v", err)
	}

	log.Println("Balance:")
	log.Printf("%d \n", balance)

	rows, err := db.Query("SELECT id, amount_cents, description FROM transactions WHERE account = $1", mockedAccountID)
	if err != nil {
		log.Fatalf("Could not retrieve transactions list: %v", err)
	}

	log.Println("Transactions:")

	for rows.Next() {
		var amt int64
		var id, desc string
		if err := rows.Scan(&id, &amt, &desc); err != nil {
			log.Fatalf("Could not scan row: %v", err)
		}

		log.Printf("%v - %v - %v\n", id, amt, desc)
	}
}
