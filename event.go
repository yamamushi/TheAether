package main

import (
	"sync"
)

// EventsDB struct
type EventsDB struct {
	db          *DBHandler
	querylocker sync.RWMutex
}

// Event struct
type Event struct {
	ID string `json:"id"`

	Name        string `json:"name"`
	Description string `json:"description"` // 60 characters or less

	Type      string   `json:"type"`
	TypeFlags []string `json:"typeflags"`

	PrivateResponse bool `json:"privateresponse"` // Whether or not to return a response in a private message
	Watchable       bool `json:"watchable"`       // Whether or not this event should be watched or just executed with a passthrough
	// If it's a passthrough, we want to write the response to the keyvaluesdb
	FinalizeOutput bool `json:"finalizeoutput"` // If set to true, we want to notify the keyvalue that our output is finalized

	LoadOnBoot bool `json:"-"` // Whether or not to load the event at boot
	//	Cycles     int      `json:"cycles"`     // Number of times to run the event, a setting of 0 or less will be parsed as "infinite"
	Data []string `json:"data"` // Different types can contain multiple data fields

	// Set when event is registered
	CreatorID string `json:"creatorid"` // The userID of the creator
	// These are not set by "events add", these must be set with the script manager
	Rooms          []string `json:"rooms"`
	ParentID       string   `json:"-"` // The id of the parent event if one exists
	ChildIDs       []string `json:"-"` // The ids of the various childs (there can exist multiple children, ie for a multiple choice question)
	RunCount       int      `json:"-"` // The total number of runs the event has had during this cycle
	TriggeredEvent bool     `json:"-"` // Used to denote whether or not an event is a copy created by a trigger

	// Used for scripting
	LinkedEvent bool `json:"-"` // If we are linked, we want to read data from the keyvalue and not passthrough data

	IsScriptEvent bool   `json:"-"` // If set to true, this event belongs to a script and should not be manually modified
	OriginalID    string `json:"originalid"`
}

// SaveEventToDB function
func (h *EventsDB) SaveEventToDB(event Event) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Events")
	err = db.Save(&event)
	return err
}

// RemoveEventFromDB function
func (h *EventsDB) RemoveEventFromDB(event Event) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Events")
	err = db.DeleteStruct(&event)
	return err
}

// RemoveEventByID function
func (h *EventsDB) RemoveEventByID(eventID string) (err error) {

	event, err := h.GetEventByID(eventID)
	if err != nil {
		return err
	}

	err = h.RemoveEventFromDB(event)
	if err != nil {
		return err
	}

	return nil
}

// GetEventByID function
func (h *EventsDB) GetEventByID(eventID string) (event Event, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Events")
	err = db.One("ID", eventID, &event)
	if err != nil {
		return event, err
	}
	return event, nil
}

// GetEventByName function
func (h *EventsDB) GetEventByName(eventName string) (event Event, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Events")
	err = db.One("Name", eventName, &event)
	if err != nil {
		return event, err
	}
	return event, nil
}

// ValidateEventByID function
func (h *EventsDB) ValidateEventByID(eventID string) (validated bool) {
	events, err := h.GetAllEvents()
	if err != nil {
		return false
	}

	for _, record := range events {

		if eventID == record.ID {
			return true
		}
	}
	return false
}

// GetAllEvents function
func (h *EventsDB) GetAllEvents() (eventlist []Event, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Events")
	err = db.All(&eventlist)
	if err != nil {
		return eventlist, err
	}
	return eventlist, nil
}
