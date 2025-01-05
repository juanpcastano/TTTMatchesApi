package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

var matchesCreated int = 0

type matchData struct {
	id            int
	movementsCode int
	winner        string
}

type WorkerPool struct {
	maxWorkers  int
	jobs        chan matchData
	batchSize   int
	maxWaitTime time.Duration
	wg          sync.WaitGroup
	db          *sql.DB
}

func NewWorkerPool(maxWorkers, batchSize int, maxWaitTime time.Duration, db *sql.DB) *WorkerPool {
	return &WorkerPool{
		maxWorkers:  maxWorkers,
		batchSize:   batchSize,
		maxWaitTime: maxWaitTime,
		jobs:        make(chan matchData, maxWorkers*batchSize),
		db:          db,
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.maxWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()

	batch := make([]matchData, 0, wp.batchSize)
	timer := time.NewTimer(wp.maxWaitTime)

	for {
		select {
		case job, ok := <-wp.jobs:
			if !ok {
				if len(batch) > 0 {
					wp.insertMatches(batch)
				}
				return
			}

			batch = append(batch, job)
			if len(batch) >= wp.batchSize {
				wp.insertMatches(batch)
				batch = make([]matchData, 0, wp.batchSize)
				timer.Reset(wp.maxWaitTime)
			}

		case <-timer.C:
			if len(batch) > 0 {
				wp.insertMatches(batch)
				batch = make([]matchData, 0, wp.batchSize)
			}
			timer.Reset(wp.maxWaitTime)
		}
	}
}

func (wp *WorkerPool) Wait() {
	close(wp.jobs)
	wp.wg.Wait()
}

func (wp *WorkerPool) AddJob(match matchData) {
	wp.jobs <- match
}

func (wp *WorkerPool) insertMatches(batch []matchData) {
	if len(batch) == 0 {
		return
	}
	query := "INSERT INTO matches (id, movementscode, winner) VALUES "
	vals := []interface{}{}
	for i, match := range batch {
		if i > 0 {
			query += ","
		}
		query += fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3)
		vals = append(vals, match.id, match.movementsCode, match.winner)
	}
	_, err := wp.db.Exec(query, vals...)
	if err != nil {
		log.Printf("Error insertando lote: %v", err)
		return
	}
}

func deleteDB() error {
	db, err := ConectToDB()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec("DELETE FROM matches")
	if err != nil {
		return err
	}
	_, err = db.Exec("ALTER SEQUENCE matches_id_seq RESTART WITH 1;")
	if err != nil {
		return err
	}
	return nil
}

func intToWinner(n int) string {
	if n == 1 {
		return "X"
	} else if n == 2 {
		return "O"
	} else {
		return ""
	}
}
func turnToInt(turn string) int {
	if turn == "X" {
		return 1
	} else if turn == "O" {
		return 2
	} else {
		log.Fatal("turn is not X or O")
		return 0
	}
}
func evaluateWinner(state []int) (bool, string) {

	isWinner := false
	winner := ""
	if state[0] != 0 {
		if state[0] == state[1] && state[1] == state[2] {
			return true, intToWinner(state[0])
		}
		if state[0] == state[4] && state[4] == state[8] {
			return true, intToWinner(state[0])
		}
		if state[0] == state[3] && state[3] == state[6] {
			return true, intToWinner(state[0])
		}
	}
	if state[4] != 0 {
		if state[4] == state[1] && state[4] == state[7] {
			return true, intToWinner(state[4])
		}
		if state[4] == state[2] && state[4] == state[6] {
			return true, intToWinner(state[4])
		}
		if state[4] == state[3] && state[4] == state[5] {
			return true, intToWinner(state[4])
		}
	}
	if state[8] != 0 {
		if state[8] == state[5] && state[5] == state[2] {
			return true, intToWinner(state[8])
		}
		if state[8] == state[7] && state[7] == state[6] {
			return true, intToWinner(state[8])
		}
	}
	return isWinner, winner
}

func writeRemainings(remainings []int, turn string, state []int, movements []int, wp *WorkerPool) {
	if len(movements) >= 9 {
		matchesCreated++
		wp.AddJob(matchData{matchesCreated, ArrayToInt(movements), "D"})
		return
	}
	for _, v := range remainings {
		for _, w := range remainings {
			state[w-1] = 0
		}
		movements = FilterSlice(movements, remainings)
		movements = append(movements, v)
		state[v-1] = turnToInt(turn)
		isWinner, winner := evaluateWinner(state)
		if isWinner {
			matchesCreated++
			wp.AddJob(matchData{matchesCreated, ArrayToInt(movements), winner})
			state[v-1] = 0
			continue
		}
		if turn == "X" {
			writeRemainings(RemoveValue(remainings, v), "O", state, movements, wp)
		} else {
			writeRemainings(RemoveValue(remainings, v), "X", state, movements, wp)
		}
	}
}

func RewriteDatabase() (int, error) {
	matchesCreated = 0
	err := deleteDB()
	if err != nil {
		return 0, err
	}

	db, err := ConectToDB()
	if err != nil {
		return 0, err
	}
	defer db.Close()
	maxWorkers, err := strconv.Atoi(os.Getenv("MAX_WORKERS"))
	if err != nil {
		return 0, err
	}
	batchSize, err := strconv.Atoi(os.Getenv("BATCH_SIZE"))
	if err != nil {
		return 0, err
	}
	maxWaitTime, err := strconv.Atoi(os.Getenv("WAIT_TIME_MS"))
	if err != nil {
		return 0, err
	}

	wp := NewWorkerPool(maxWorkers, batchSize, time.Duration(maxWaitTime)*time.Millisecond, db)
	wp.Start()

	remainings := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	initialState := make([]int, 9)
	movements := []int{}

	writeRemainings(remainings, "X", initialState, movements, wp)

	wp.Wait()

	return matchesCreated, nil
}
