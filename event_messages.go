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
	ID string `storm:"id"`

	ScriptID       string
	EventsComplete bool

	// An event message can contain many types
	// Rather than making this archaic and having to parse out what we mean by a given response
	// We are creating several different values and types here to be parsed easier later on
	// i.e. a die roll should go into "Roll" rather than a generic int variable
	Roll int

	CheckError   bool
	ErrorMessage string

	CheckSuccess bool
	Successful   bool
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

// ClearEventMessage function
func (h *EventMessagesDB) ClearEventMessage(eventmessageID string, scriptID string) (err error) {

	err = h.RemoveEventMessageByID(eventmessageID)
	if err != nil {
		return err
	}

	eventmessage := EventMessageContainer{ID: eventmessageID, ScriptID: scriptID}
	err = h.SaveEventMessageToDB(eventmessage)
	if err != nil {
		return err
	}

	return nil
}

// TerminateEvents function
func (h *EventMessagesDB) TerminateEvents(eventmessageID string) (err error) {
	eventmessage, err := h.GetEventMessageByID(eventmessageID)
	if err != nil {
		return err
	}
	eventmessage.EventsComplete = true
	err = h.SaveEventMessageToDB(eventmessage)
	if err != nil {
		return err
	}
	return nil
}

// SetSuccessfulStatus function
func (h *EventMessagesDB) SetSuccessfulStatus(eventmessageID string) (err error) {
	eventmessage, err := h.GetEventMessageByID(eventmessageID)
	if err != nil {
		return err
	}
	eventmessage.CheckSuccess = true
	eventmessage.Successful = true
	err = h.SaveEventMessageToDB(eventmessage)
	if err != nil {
		return err
	}
	return nil
}

// SetFailureStatus function
func (h *EventMessagesDB) SetFailureStatus(eventmessageID string) (err error) {
	eventmessage, err := h.GetEventMessageByID(eventmessageID)
	if err != nil {
		return err
	}
	eventmessage.CheckSuccess = true
	eventmessage.Successful = false
	err = h.SaveEventMessageToDB(eventmessage)
	if err != nil {
		return err
	}
	return nil
}

// SetErrorMessage function
func (h *EventMessagesDB) SetErrorMessage(eventmessageID string, message string) (err error) {
	eventmessage, err := h.GetEventMessageByID(eventmessageID)
	if err != nil {
		return err
	}
	eventmessage.CheckError = true
	eventmessage.ErrorMessage = message
	err = h.SaveEventMessageToDB(eventmessage)
	if err != nil {
		return err
	}
	return nil
}
