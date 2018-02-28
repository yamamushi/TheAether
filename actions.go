package main

import (
	"sync"
)

// ActionsDB struct 
type ActionsDB struct {
	db          *DBHandler
	querylocker sync.RWMutex
}

// Action struct
type Action struct {
	ID   string `json:"id"`
	Name string

	Type string
	/*
	   There are six types of actions:

	   1)Standard
	   2)Move
	   3)Full-round
	   4)Swift
	   5)Immediate
	   6)Free
	*/

	// Actions may have an attribute requirement associated with them
	// However equipment, spells, items, etc may all give attribute check bonuses too
	Strength     int
	Dexterity    int
	Constitution int
	Intelligence int
	Wisdom       int
	Charisma     int

	Description string // A brief description about the action

	SkillRequirement string // There may be a skill associated with an action, not all actions will have these
	Timeout          int    // Time in seconds before using this action again
}

// SaveActionToDB function
func (h *ActionsDB) SaveActionToDB(action Action) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Actions")
	err = db.Save(&action)
	return err
}

// RemoveActionFromDB function
func (h *ActionsDB) RemoveActionFromDB(action Action) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Actions")
	err = db.DeleteStruct(&action)
	return err
}

// RemoveActionByID function
func (h *ActionsDB) RemoveActionByID(actionID string) (err error) {

	action, err := h.GetActionByID(actionID)
	if err != nil {
		return err
	}

	err = h.RemoveActionFromDB(action)
	if err != nil {
		return err
	}
	return nil
}

// GetActionByID function
func (h *ActionsDB) GetActionByID(actionID string) (action Action, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Actions")
	err = db.One("ID", actionID, &action)
	if err != nil {
		return action, err
	}
	return action, nil
}

// GetActionByName function
func (h *ActionsDB) GetActionByName(actionName string) (action Action, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Actions")
	err = db.One("Name", actionName, &action)
	if err != nil {
		return action, err
	}
	return action, nil
}

// ValidateActionByID function
func (h *ActionsDB) ValidateActionByID(actionID string) (validated bool) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	var action Action
	db := h.db.rawdb.From("Actions")
	err := db.One("ID", actionID, &action)
	if err != nil {
		return false
	}
	return true
}

// GetAllActions function
func (h *ActionsDB) GetAllActions() (actionlist []Action, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Actions")
	err = db.All(&actionlist)
	if err != nil {
		return actionlist, err
	}
	return actionlist, nil
}
