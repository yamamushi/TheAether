package main

import "sync"

// A general purpose key-value mechanism that uses the database for storing temporary key value states from
// This database should be flushed every time the server launches.

// EventMessagesDB struct
type EventMessagesDB struct {
	db          *DBHandler
	querylocker sync.RWMutex
}

// EventMessageContainer struct
type EventMessageContainer struct {
	ID string

	EventsComplete bool

	// An event message can contain many types
	// Rather than making this archaic and having to parse out what we mean by a given response
	// We are creating several different values and types here to be parsed easier later on
	// i.e. a die roll should go into "Roll" rather than a generic int variable
	Roll       int
	Response   string
	Successful bool
}

// SaveEventMessageToDB function
func (h *EventMessagesDB) SaveEventMessageToDB(eventMessage EventMessageContainer) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("EventMessages")
	err = db.Save(&eventMessage)
	return err
}

// RemoveEventMessageFromDB function
func (h *EventMessagesDB) RemoveEventMessageFromDB(eventMessage EventMessageContainer) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("EventMessages")
	err = db.DeleteStruct(&eventMessage)
	return err
}

// RemoveEventMessageByID function
func (h *EventMessagesDB) RemoveEventMessageByID(eventMessageID string) (err error) {

	eventMessage, err := h.GetEventMessageByID(eventMessageID)
	if err != nil {
		return err
	}

	err = h.RemoveEventMessageFromDB(eventMessage)
	if err != nil {
		return err
	}
	return nil
}

// GetEventMessageByID function
func (h *EventMessagesDB) GetEventMessageByID(eventMessageID string) (eventMessage EventMessageContainer, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("EventMessages")
	err = db.One("ID", eventMessageID, &eventMessage)
	if err != nil {
		return eventMessage, err
	}
	return eventMessage, nil
}

// GetAllEventMessages function
func (h *EventMessagesDB) GetAllEventMessages() (eventMessageList []EventMessageContainer, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("EventMessages")
	err = db.All(&eventMessageList)
	if err != nil {
		return eventMessageList, err
	}
	return eventMessageList, nil
}

// Init function
func (h *EventMessagesDB) Init() (err error) {
	err = h.db.rawdb.Drop("EventMessages")
	if err != nil {
		if err.Error() == "bucket not found" {
			return nil
		}
		return err
	}
	return nil
}
