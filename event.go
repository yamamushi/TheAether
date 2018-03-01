package main

import (
	"errors"
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

	Name         string   `json:"name" storm:"unique"` // Names must be unique
	Description  string   `json:"description"`         // 60 characters or less
	Rooms        []string `json:"rooms"`
	UserAttached string   `json:"attacheduserid"` // The ID of a user if the event is tied to one through a cycle count

	Type      string   `json:"type"`
	TypeFlags []string `json:"typeflags"`

	PrivateResponse bool `json:"privateresponse"` // Whether or not to return a response in a private message
	Attachable      bool `json:"attachable"`      // Whether or not this event can be attached to a user or not
	Watchable       bool `json:"watchable"`       // Whether or not this event should be watched or just executed with a passthrough
	// If it's a passthrough, we want to write the response to the keyvaluesdb

	LoadOnBoot bool     `json:"loadonboot"` // Whether or not to load the event at boot
	Cycles     int      `json:"cycles"`     // Number of times to run the event, a setting of 0 or less will be parsed as "infinite"
	Data       []string `json:"data"`       // Different types can contain multiple data fields

	// Set when event is registered
	CreatorID string `json:"creatorid"` // The userID of the creator
	// These are not set by "events add", these must be set with the script manager
	ParentID       string   `json:"parentid"`       // The id of the parent event if one exists
	ChildIDs       []string `json:"childids"`       // The ids of the various childs (there can exist multiple children, ie for a multiple choice question)
	RunCount       int      `json:"runcount"`       // The total number of runs the event has had during this cycle
	TriggeredEvent bool     `json:"triggeredevent"` // Used to denote whether or not an event is a copy created by a trigger
	KeyValueID     string   `json:"keyvalueid"`     // If we are writing to a keyvalue, we need to know the ID to write to
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

// GetEventByAttached function
func (h *EventsDB) GetEventByAttached(eventID string, UserID string) (event Event, err error) {
	events, err := h.GetAllEvents()
	if err != nil {
		return event, err
	}

	searchstring := eventID + "-" + UserID
	for _, record := range events {
		if searchstring == record.UserAttached {
			return record, nil
		}
	}
	return event, errors.New("No record found")
}
