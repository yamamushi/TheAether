package main

import (

	"sync"
	"errors"
)


type EventDB struct {
	db          *DBHandler
	querylocker sync.RWMutex
}

type Event struct {

	ID			string

	Type		string
	TypeFlags	[]string

	TimeDelay	string // We parse this into a duration

	Data		string

}



func (h *EventDB) SaveEventToDB(Event Event) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Events")
	err = db.Save(&Event)
	return err
}

func (h *EventDB) RemoveEventFromDB(Event Event) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Events")
	err = db.DeleteStruct(&Event)
	return err
}

func (h *EventDB) RemoveEventByID(EventID string) (err error) {

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

func (h *EventDB) GetEventByID(EventID string) (Event Event, err error) {

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
func (h *EventDB) GetAllEvents() (Eventlist []Event, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("Events")
	err = db.All(&Eventlist)
	if err != nil {
		return Eventlist, err
	}

	return Eventlist, nil
}



