package storage

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

func lcp(first string, second string) int {
	i := 0
	for i < len(first) && i < len(second) && first[i] == second[i] {
		i++
	}
	return i
}

func isStreamEntryIdValid(id string, lastSavedId string) (string, error) {
	lastSavedParts := strings.Split(lastSavedId, "-")
	if len(lastSavedParts) != 2 {
		return "", fmt.Errorf("invalid id saved as the last entry. lastSavedId: %s", lastSavedId)
	}
	lastSavedMil := lastSavedParts[0]
	lastSavedSeq := lastSavedParts[1]

	if id == "*" {
		currentMil := time.Now().UnixMilli()
		newId := fmt.Sprintf("%d-0", currentMil)
		return newId, nil
	}

	newIdParts := strings.Split(id, "-")
	if len(newIdParts) != 2 {
		return "", fmt.Errorf("invalid id string for new entry. id: %s", id)
	}
	newIdMil := newIdParts[0]
	newIdSeq := newIdParts[1]

	if _, err := strconv.Atoi(newIdMil); err != nil {
		return "", fmt.Errorf("left part of id should be a number. id: %s", id)
	}

	if newIdMil == "0" && newIdSeq == "0" {
		return "", errors.New("The ID specified in XADD must be greater than 0-0")
	}

	if lastSavedMil > newIdMil {
		return "", errors.New("The ID specified in XADD is equal or smaller than the target stream top item")
	}

	if newIdSeq == "*" {
		intLastSavedSeq, err := strconv.Atoi(lastSavedSeq)
		if err != nil {
			return "", fmt.Errorf("error converting right part of last saved id. id seq: %s", lastSavedSeq)
		}
		if lastSavedMil == newIdMil {
			id = fmt.Sprintf("%s-%d", lastSavedMil, intLastSavedSeq+1)
		} else {
			id = fmt.Sprintf("%s-%s", newIdMil, "0")
		}

	} else {
		if newIdMil == lastSavedMil && lastSavedSeq >= newIdSeq {
			return "", errors.New("The ID specified in XADD is equal or smaller than the target stream top item")
		}
	}

	return id, nil
}

func validateXRangeIds(startId string, endId string) (string, string) {
	if !strings.Contains(startId, "-") {
		startId = fmt.Sprintf("%s-0", startId)
	}
	if !strings.Contains(endId, "-") {
		endId = fmt.Sprintf("%s-%d", endId, math.MaxInt32)
	}
	return startId, endId
}

type comparisonMode string

const (
	lowerEqual   comparisonMode = "lower"
	greaterEqual comparisonMode = "greater"
)

func compareStreamIds(currentId string, limitId string, mode comparisonMode) bool {
	currentIdParts := strings.Split(currentId, "-")
	currentIdMil, _ := strconv.Atoi(currentIdParts[0])
	currentIdSeq, _ := strconv.Atoi(currentIdParts[1])

	limitIdParts := strings.Split(limitId, "-")
	limitIdMil, _ := strconv.Atoi(limitIdParts[0])
	limitIdSeq, _ := strconv.Atoi(limitIdParts[1])

	if mode == lowerEqual {
		if currentIdMil == limitIdMil {
			return currentIdSeq <= limitIdSeq
		}
		return currentIdMil <= limitIdMil
	}

	if currentIdMil == limitIdMil {
		return currentIdSeq >= limitIdSeq
	}
	return currentIdMil >= limitIdMil
}
