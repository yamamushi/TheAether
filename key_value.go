package main

import "sync"

// A general purpose key-value mechanism that uses the database for storing temporary key value states from
// This database should be flushed every time the server launches.

// KeyValuesDB struct
type KeyValuesDB struct {
	db          *DBHandler
	querylocker sync.RWMutex
}

// KeyValue struct
type KeyValue struct {
	ID string

	Passed   bool
	Value    int
	Response string
}

// SaveKeyValueToDB function
func (h *KeyValuesDB) SaveKeyValueToDB(keyValue KeyValue) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("KeyValues")
	err = db.Save(&keyValue)
	return err
}

// RemoveKeyValueFromDB function
func (h *KeyValuesDB) RemoveKeyValueFromDB(keyValue KeyValue) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("KeyValues")
	err = db.DeleteStruct(&keyValue)
	return err
}

// RemoveKeyValueByID function
func (h *KeyValuesDB) RemoveKeyValueByID(keyValueID string) (err error) {

	keyValue, err := h.GetKeyValueByID(keyValueID)
	if err != nil {
		return err
	}

	err = h.RemoveKeyValueFromDB(keyValue)
	if err != nil {
		return err
	}
	return nil
}

// GetKeyValueByID function
func (h *KeyValuesDB) GetKeyValueByID(keyValueID string) (keyValue KeyValue, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("KeyValues")
	err = db.One("ID", keyValueID, &keyValue)
	if err != nil {
		return keyValue, err
	}
	return keyValue, nil
}

// GetAllKeyValues function
func (h *KeyValuesDB) GetAllKeyValues() (keyValuelist []KeyValue, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("KeyValues")
	err = db.All(&keyValuelist)
	if err != nil {
		return keyValuelist, err
	}
	return keyValuelist, nil
}
