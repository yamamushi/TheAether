package main

import (
	"github.com/satori/go.uuid"
)

// GetUUID function
func GetUUID() (id string, err error) {

	formattedid, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	return formattedid.String(), nil

}

// GetUUIDv2 function
// Ignore Errors with this
func GetUUIDv2() (id string){
	id, _ = GetUUID()
	return id
}