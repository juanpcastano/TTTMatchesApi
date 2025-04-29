package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var endpointsArray []string = []string{"/info", "/matchByID/{id}", "/matchByMC/{MovementsCode}"}

func LoadDotEnv() {
	err := godotenv.Load()
	if err != nil {
		panic(fmt.Sprintf("error loading .env: %v\n", err))
	}
}

type dbInfo struct {
	TotalMatches    int      `json:"total_matches"`
	XWinningMatches int      `json:"x_winning_matches"`
	OWinningMatches int      `json:"o_winning_matches"`
	Draws           int      `json:"draws"`
	Endpoints       []string `json:"endpoints"`
}

type Match struct {
	Id            int        `json:"id"`
	MovementsCode int        `json:"movements_code"`
	Winner        string     `json:"winner"`
	Movements     []Movement `json:"movements"`
}

type Movement struct {
	MovementNumber int    `json:"movement_number"`
	IsWinner       bool   `json:"is_winner"`
	StateCode      string `json:"state_code"`
}

func info(w http.ResponseWriter, req *http.Request) {
	db, err := ConectToDB()
	if err != nil {
		fmt.Printf("err: %v\n", err)
		WriteJSON(w, 500, map[string]string{"error": "could not connect to database"})
		return
	}
	defer db.Close()

	dbInfo := dbInfo{
		Endpoints: endpointsArray,
	}

	err = db.QueryRow("SELECT count(*) FROM matches;").Scan(&dbInfo.TotalMatches)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		WriteJSON(w, 500, map[string]string{"error": "query failed"})
		return
	}

	err = db.QueryRow("SELECT count(*) FROM matches WHERE winner = 'X';").Scan(&dbInfo.XWinningMatches)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		WriteJSON(w, 500, map[string]string{"error": "query failed"})
		return
	}

	err = db.QueryRow("SELECT count(*) FROM matches WHERE winner = 'O';").Scan(&dbInfo.OWinningMatches)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		WriteJSON(w, 500, map[string]string{"error": "query failed"})
		return
	}

	err = db.QueryRow("SELECT count(*) FROM matches WHERE winner = 'D';").Scan(&dbInfo.Draws)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		WriteJSON(w, 500, map[string]string{"error": "query failed"})
		return
	}

	WriteJSON(w, 200, dbInfo)
}

func matchByID(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]

	if _, err := strconv.Atoi(id); err != nil {
		WriteJSON(w, 400, map[string]string{"error": "Invalid match ID"})
		return
	}

	match := &Match{}
	db, err := ConectToDB()
	if err != nil {
		fmt.Printf("Error connecting to DB: %v\n", err)
		WriteJSON(w, 500, map[string]string{"error": "could not connect to database"})
		return
	}
	defer db.Close()

	queryMatch := "SELECT * FROM matches WHERE id = $1"
	err = db.QueryRow(queryMatch, id).Scan(&match.Id, &match.MovementsCode, &match.Winner)
	if err != nil {
		if err == sql.ErrNoRows {
			WriteJSON(w, 404, map[string]string{"error": "Match not found"})
			return
		}
		fmt.Printf("Error fetching match: %v\n", err)
		WriteJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	match.Movements = MovementsCodeToMovementsArray(strconv.Itoa(match.MovementsCode))

	WriteJSON(w, 200, match)
}

func matchByMC(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	mc := vars["MovementsCode"]
	if _, err := strconv.Atoi(mc); err != nil {
		WriteJSON(w, 400, map[string]string{"error": "Invalid match MC"})
		return
	}
	var match Match
	db, err := ConectToDB()
	if err != nil {
		fmt.Printf("err: %v\n", err)
		WriteJSON(w, 500, map[string]string{"error": "could not connect to database"})
		return
	}
	defer db.Close()
	err = db.QueryRow("SELECT * FROM matches WHERE movementscode = $1", mc).Scan(&match.Id, &match.MovementsCode, &match.Winner)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			WriteJSON(w, 200, map[string]string{"error": "Match not found"})
			return
		}
		fmt.Printf("err: %v\n", err)
		WriteJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	match.Movements = MovementsCodeToMovementsArray(strconv.Itoa(match.MovementsCode))

	WriteJSON(w, 200, match)

}

func rewriteDatabase(w http.ResponseWriter, req *http.Request) {
	startTime := time.Now()
	params := req.URL.Query()
	if params.Get("password") != os.Getenv("API_PASSWORD") {
		WriteJSON(w, 401, map[string]string{"error": "not autorized"})
		return
	}

	matchesCreated, err := RewriteDatabase()
	if err != nil {
		fmt.Printf("err: %v\n", err)
		WriteJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	duration := time.Since(startTime)

	WriteJSON(w, 200, map[string]interface{}{
		"status":                 "succesfull",
		"matches_created":        matchesCreated,
		"execution_time":         duration.String(),
		"execution_time_seconds": duration.Seconds(),
	})
}

func main() {
	LoadDotEnv()
	port := ":" + os.Getenv("PORT")
	server := mux.NewRouter()
	server.HandleFunc("/info", info).Methods("GET")
	server.HandleFunc("/matchByID/{id}", matchByID).Methods("GET")
	server.HandleFunc("/matchByMC/{MovementsCode}", matchByMC).Methods("GET")
	server.HandleFunc("/rewriteDatabase", rewriteDatabase).Methods("GET")
	fmt.Println("Serving On Port", port[1:])
	err := http.ListenAndServe(port, server)
	if err != nil {
		fmt.Printf("Error Starting The Server: %v\n", err)
	}
}
