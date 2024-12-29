package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

const port string = ":8090"

var endpointsArray []string = []string{"/info", "/matchByID/{id}", "/matchByMC/{MovementsCode}"}

func writeJSON(w http.ResponseWriter, status int, value any) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(value)
}

type dbInfo struct {
	TotalMatches    int      `json:"total_matches"`
	XWinningMatches int      `json:"x_winning_matches"`
	OWinningMatches int      `json:"o_winning_matches"`
	Draws           int      `json:"draws"`
	Endpoints       []string `json:"endpoints"`
}

type match struct {
	Id            int        `json:"id"`
	MovementsCode int        `json:"movements_code"`
	Winner        string     `json:"winner"`
	Movements     []movement `json:"movements"`
}

type movement struct {
	MovementNumber int  `json:"movement_number"`
	IsWinner       bool `json:"is_winner"`
	StateCode      int  `json:"state_code"`
}

func info(w http.ResponseWriter, req *http.Request) {
	db, err := ConectToDB()
	if err != nil {
		fmt.Printf("err: %v\n", err)
		writeJSON(w, 500, map[string]string{"error": "could not connect to database"})
		return
	}
	defer db.Close()

	dbInfo := dbInfo{
		Endpoints: endpointsArray,
	}

	err = db.QueryRow("SELECT count(*) FROM matches;").Scan(&dbInfo.TotalMatches)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		writeJSON(w, 500, map[string]string{"error": "query failed"})
		return
	}

	err = db.QueryRow("SELECT count(*) FROM matches WHERE winner = 'X';").Scan(&dbInfo.XWinningMatches)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		writeJSON(w, 500, map[string]string{"error": "query failed"})
		return
	}

	err = db.QueryRow("SELECT count(*) FROM matches WHERE winner = 'O';").Scan(&dbInfo.OWinningMatches)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		writeJSON(w, 500, map[string]string{"error": "query failed"})
		return
	}

	err = db.QueryRow("SELECT count(*) FROM matches WHERE winner = 'D';").Scan(&dbInfo.Draws)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		writeJSON(w, 500, map[string]string{"error": "query failed"})
		return
	}

	writeJSON(w, 200, dbInfo)
}

func matchByID(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]

	if _, err := strconv.Atoi(id); err != nil {
		writeJSON(w, 400, map[string]string{"error": "Invalid match ID"})
		return
	}

	match := &match{}
	db, err := ConectToDB()
	if err != nil {
		fmt.Printf("Error connecting to DB: %v\n", err)
		writeJSON(w, 500, map[string]string{"error": "could not connect to database"})
		return
	}
	defer db.Close()

	queryMatch := "SELECT * FROM matches WHERE id = $1"
	err = db.QueryRow(queryMatch, id).Scan(&match.Id, &match.MovementsCode, &match.Winner)
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSON(w, 404, map[string]string{"error": "Match not found"})
			return
		}
		fmt.Printf("Error fetching match: %v\n", err)
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	rows, err := db.Query("SELECT movementnumber, iswinner, statecode FROM movements WHERE match_id = $1", id)
	if err != nil {
		fmt.Printf("Error fetching movements: %v\n", err)
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		movementOBJ := &movement{}
		if err := rows.Scan(&movementOBJ.MovementNumber, &movementOBJ.IsWinner, &movementOBJ.StateCode); err != nil {
			fmt.Printf("Error scanning row: %v\n", err)
			continue
		}
		match.Movements = append(match.Movements, *movementOBJ)
	}

	writeJSON(w, 200, match)
}

func matchByMC(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	mc := vars["MovementsCode"]
	if _, err := strconv.Atoi(mc); err != nil {
		writeJSON(w, 400, map[string]string{"error": "Invalid match ID"})
		return
	}
	var match match
	db, err := ConectToDB()
	if err != nil {
		fmt.Printf("err: %v\n", err)
		writeJSON(w, 500, map[string]string{"error": "could not connect to database"})
		return
	}
	defer db.Close()
	err = db.QueryRow("SELECT * FROM matches WHERE movementscode = $1", mc).Scan(&match.Id, &match.MovementsCode, &match.Winner)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			writeJSON(w, 200, map[string]string{"error": "Match not found"})
			return
		}
		fmt.Printf("err: %v\n", err)
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, 200, match)

}

func main() {
	server := mux.NewRouter()
	server.HandleFunc("/info", info).Methods("GET")
	server.HandleFunc("/matchByID/{id}", matchByID).Methods("GET")
	server.HandleFunc("/matchByMC/{MovementsCode}", matchByMC).Methods("GET")
	fmt.Println("Serving On Port", port[1:])
	err := http.ListenAndServe(port, server)
	if err != nil {
		fmt.Printf("Error Starting The Server: %v\n", err)
	}
}