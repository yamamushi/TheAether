package main

import "sync"

// ScriptsDB struct
type ScriptsDB struct {
	db          *DBHandler
	querylocker sync.RWMutex
}

// Script struct
type Script struct {
	ID string `storm:"id"`

	Name        string `storm:"unique"`
	Description string

	Executable bool

	CreatorID       string
	EventIDs        []string // The sequential list of events that comprise a script
	Synchronized    bool
	EventMessagesID string
}

// SaveScriptToDB function
func (h *ScriptsDB) SaveScriptToDB(script Script) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Scripts")
	err = db.Save(&script)
	return err
}

// RemoveScriptFromDB function
func (h *ScriptsDB) RemoveScriptFromDB(script Script) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Scripts")
	err = db.DeleteStruct(&script)
	return err
}

// RemoveScriptByID function
func (h *ScriptsDB) RemoveScriptByID(scriptID string) (err error) {
	script, err := h.GetScriptByID(scriptID)
	if err != nil {
		return err
	}

	err = h.RemoveScriptFromDB(script)
	if err != nil {
		return err
	}
	return nil
}

// GetScriptByID function
func (h *ScriptsDB) GetScriptByID(scriptID string) (script Script, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Scripts")
	err = db.One("ID", scriptID, &script)
	if err != nil {
		return script, err
	}
	return script, nil
}

// GetScriptByName function
func (h *ScriptsDB) GetScriptByName(scriptName string) (script Script, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Scripts")
	err = db.One("Name", scriptName, &script)
	if err != nil {
		return script, err
	}
	return script, nil
}

// GetAllScripts function
func (h *ScriptsDB) GetAllScripts() (scriptlist []Script, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Scripts")
	err = db.All(&scriptlist)
	if err != nil {
		return scriptlist, err
	}
	return scriptlist, nil
}
