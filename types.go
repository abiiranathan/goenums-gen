// Code generated by "goenums"; DO NOT EDIT.

package main

import (
	"database/sql/driver"
	"fmt"
)

type OperationStatus string

const (
	OperationStatusPending   OperationStatus = "PENDING"
	OperationStatusOnGoing   OperationStatus = "ON GOING"
	OperationStatusCompleted OperationStatus = "COMPLETED"
	OperationStatusPostponed OperationStatus = "POSTPONED"
	OperationStatusCancelled OperationStatus = "CANCELLED"
)

func (e OperationStatus) IsValid() bool {
	validValues := []string{
		"PENDING",
		"ON GOING",
		"COMPLETED",
		"POSTPONED",
		"CANCELLED",
	}

	for _, val := range validValues {
		if val == string(e) {
			return true
		}
	}
	return false
}

func (e OperationStatus) ValidValues() []string {
	return []string{
		"PENDING",
		"ON GOING",
		"COMPLETED",
		"POSTPONED",
		"CANCELLED",
	}
}

func (e *OperationStatus) Scan(src interface{}) error {
	source, ok := src.(string)
	if !ok {
		return fmt.Errorf("invalid value for %s: %s", "OperationStatus", source)
	}
	*e = OperationStatus(source)
	return nil
}

func (e OperationStatus) Value() (driver.Value, error) {
	if !e.IsValid() {
		return nil, fmt.Errorf("invalid value for %s", "OperationStatus")
	}
	return string(e), nil
}
