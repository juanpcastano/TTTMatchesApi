package main

import "strconv"

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
