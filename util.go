package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func WriteJSON(w http.ResponseWriter, status int, value any) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(value)
}

func MovementsCodeToMovementsArray(mc string) []Movement {
	stateCode := []byte("---------")
	result := []Movement{}

	for i := range len(mc) {
		player := 'X'
		if i%2 != 0 {
			player = 'O'
		}
		index, err := strconv.Atoi(string(mc[i]))
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
		stateCode[index-1] = byte(player)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
		isWinner := false
		if i+1 == len(mc) {
			isWinner = true
		}
		result = append(result, Movement{
			MovementNumber: i + 1,
			IsWinner:       isWinner,
			StateCode:      string(stateCode),
		})
	}
	return result
}

func FilterSlice(original, toRemove []int) []int {
	removeMap := make(map[int]bool)
	for _, val := range toRemove {
		removeMap[val] = true
	}

	result := []int{}
	for _, val := range original {
		if !removeMap[val] {
			result = append(result, val)
		}
	}

	return result
}

func ArrayToInt(arr []int) int {
	resultStr := ""
	for _, digit := range arr {
		resultStr += strconv.Itoa(digit)
	}
	resultInt, _ := strconv.Atoi(resultStr)
	return resultInt
}

func RemoveValue(slice []int, value int) []int {
	result := []int{}
	for _, v := range slice {
		if v != value {
			result = append(result, v)
		}
	}
	return result
}
