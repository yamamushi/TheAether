package main

import (

	"sync"
	"errors"
)

// EventsDB struct
type EventsDB struct {
	db          *DBHandler
	querylocker sync.RWMutex
}

// Event struct
type Event struct {

	ID			string `json:"id"`

	ChannelID  	string `json:"channelid"`

	Type		string `json:"type"`
	TypeFlags	[]string `json:"typeflags"`

	CreatorID 	string `json:"creatorid"`// The userID of the creator
	Data		string `json:"data"`

}


// SaveEventToDB function
func (h *EventsDB) SaveEventToDB(Event Event) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Events")
	err = db.Save(&Event)
	return err
}

// RemoveEventFromDB function
func (h *EventsDB) RemoveEventFromDB(Event Event) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Events")
	err = db.DeleteStruct(&Event)
	return err
}

// RemoveEventByID function
func (h *EventsDB) RemoveEventByID(EventID string) (err error) {

	Event, err := h.GetEventByID(EventID)
	if err != nil {
		return err
	}

	err = h.RemoveEventFromDB(Event)
	if err != nil {
		return err
	}

	return nil
}

// GetEventByID function
func (h *EventsDB) GetEventByID(EventID string) (Event Event, err error) {

	Events, err := h.GetAllEvents()
	if err != nil{
		return Event, err
	}

	for _, record := range Events {

		if EventID == record.ID {
			return record, nil
		}
	}
	return Event, errors.New("No record found")
}


// GetAllEvents function
func (h *EventsDB) GetAllEvents() (Eventlist []Event, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Events")
	err = db.All(&Eventlist)
	if err != nil {
		return Eventlist, err
	}

	return Eventlist, nil
}



