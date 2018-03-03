package main

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

// EventParser struct
// Parse event scripts and formatted event data fields
type EventParser struct {
	eventsdb *EventsDB
}

// ParseFormattedEvent function
func (h *EventParser) ParseFormattedEvent(data string, userID string) (parsed Event, err error) {
	unmarshallcontainer := Event{}
	if err := json.Unmarshal([]byte(data), &unmarshallcontainer); err != nil {
		return unmarshallcontainer, err
	}

	unmarshallcontainer.CreatorID = userID
	unmarshallcontainer.RunCount = 0
	if unmarshallcontainer.Name == "" {
		return parsed, errors.New("Event requires a name")
	}
	if len(unmarshallcontainer.Name) > 30 {
		return parsed, errors.New("Name must not exceed 30 characters")
	}
	if unmarshallcontainer.Description == "" {
		return parsed, errors.New("Event requires a description")
	}
	if len(unmarshallcontainer.Description) > 60 {
		return parsed, errors.New("Description must not exceed 60 characters")
	}
	_, err = h.eventsdb.GetEventByName(unmarshallcontainer.Name)
	if err == nil {
		return parsed, errors.New("Event with name: " + unmarshallcontainer.Name + " already exists")
	}
	if unmarshallcontainer.IsScriptEvent {
		return parsed, errors.New("Event cannot have scriptevent defined")
	}
	if unmarshallcontainer.LoadOnBoot {
		return parsed, errors.New("Event cannot manually be loaded on boot")
	}

	// Generate and assign an ID to this event
	id := strings.Split(GetUUIDv2(), "-")
	unmarshallcontainer.ID = id[0]
	return unmarshallcontainer, nil
}

// VerifyUpdateEvent function
func (h *EventParser) VerifyUpdateEvent(data string, userID string) (parsed Event, err error) {
	unmarshallcontainer := Event{}
	if err := json.Unmarshal([]byte(data), &unmarshallcontainer); err != nil {
		return unmarshallcontainer, err
	}

	unmarshallcontainer.CreatorID = userID
	unmarshallcontainer.RunCount = 0
	if unmarshallcontainer.Name == "" {
		return parsed, errors.New("Event requires a name")
	}
	if len(unmarshallcontainer.Name) > 30 {
		return parsed, errors.New("Name must not exceed 30 characters")
	}
	if unmarshallcontainer.Description == "" {
		return parsed, errors.New("Event requires a description")
	}
	if len(unmarshallcontainer.Description) > 60 {
		return parsed, errors.New("Description must not exceed 60 characters")
	}
	if unmarshallcontainer.IsScriptEvent {
		return parsed, errors.New("Event cannot have scriptevent defined")
	}
	if unmarshallcontainer.LoadOnBoot {
		return parsed, errors.New("Event cannot manually be loaded on boot")
	}

	return unmarshallcontainer, nil
}

// EventToJSON function
func (h *EventParser) EventToJSON(event Event) (formatted string, err error) {
	marshalledevent, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	formatted = string(marshalledevent)
	return formatted, nil
}

// ValidateEvent function
// This will take an event and validate the correct number of arguments were passed to it
// Refer to the github wiki page on Events for information on types
func (h *EventParser) ValidateEvent(event Event) (err error) {
	if event.Type == "ReadMessage" {
		return h.ValidateReadMessage(event)
	} else if event.Type == "ReadTimedMessage" {
		return h.ValidateReadTimedMessage(event)
	} else if event.Type == "ReadMessageChoice" {
		return h.ValidateReadMessageChoice(event)
	} else if event.Type == "ReadMessageChoiceTriggerEvent" {
		return h.ValidateReadMessageChoiceTriggerEvent(event)
	} else if event.Type == "SendMessageEvent" {
		return h.ValidateSendMessage(event)
	} else if event.Type == "TimedSendMessageEvent" {
		return h.ValidateTimedSendMessageEvent(event)
	} else if event.Type == "MessageTriggerSuccessFail" {
		return h.ValidateMessageTriggerSuccessFail(event)
	} else if event.Type == "TriggerSuccess" {
		return h.ValidateTriggerSuccess(event)
	} else if event.Type == "TriggerFailure" {
		return h.ValidateTriggerFailure(event)
	} else if event.Type == "SendMessageTriggerEvent" {
		return h.ValidateSendMessageTriggerEvent(event)
	} else if event.Type == "TriggerFailureSendError" {
		return h.ValidateTriggerFailureSendError(event)
	} else if event.Type == "MessageChoiceDefaultEvent" {
		return h.ValidateMessageChoiceDefaultEvent(event)
	} else if event.Type == "MessageChoiceDefault" {
		return h.ValidateMessageChoiceDefaultEvent(event)
	}
	return errors.New("unrecognized event type: " + event.Type)
}

// HasEventsInData function
func (h *EventParser) HasEventsInData(event Event) (hasevents bool) {
	if event.Type == "ReadMessage" {
		return false
	} else if event.Type == "ReadTimedMessage" {
		return false
	} else if event.Type == "ReadMessageChoice" {
		return false
	} else if event.Type == "ReadMessageChoiceTriggerEvent" {
		return true
	} else if event.Type == "SendMessage" {
		return false
	} else if event.Type == "TimedSendMessage" {
		return false
	} else if event.Type == "MessageTriggerSuccessFail" {
		return false
	} else if event.Type == "TriggerSuccess" {
		return false
	} else if event.Type == "TriggerFailure" {
		return false
	} else if event.Type == "SendMessageTriggerEvent" {
		return true
	}
	return false
}

// ValidateReadMessage function
func (h *EventParser) ValidateReadMessage(event Event) (err error) {
	if len(event.TypeFlags) != 1 {
		return errors.New("Error validating event - Expected 1 typeflag but found: " + strconv.Itoa(len(event.TypeFlags)))
	}
	return nil
}

// ValidateReadTimedMessage function
func (h *EventParser) ValidateReadTimedMessage(event Event) (err error) {
	if len(event.TypeFlags) != 2 {
		return errors.New("Error validating event - Expected 2 typeflags but found: " + strconv.Itoa(len(event.TypeFlags)))
	}
	timeout, err := strconv.Atoi(event.TypeFlags[1])
	if err != nil {
		return errors.New("Error validating event - Could not parse timeout: " + err.Error())
	}
	if timeout > 300 {
		return errors.New("Error validating event - Maximum timeout is 300 but found: " + strconv.Itoa(timeout))
	}
	return nil
}

// ValidateReadMessageChoice function
func (h *EventParser) ValidateReadMessageChoice(event Event) (err error) {
	typeflagslen := len(event.TypeFlags)
	datafieldslen := len(event.Data)
	if len(event.TypeFlags) < 1 {
		return errors.New("Error validating event - Expected at least 1 typeflag")
	}
	if typeflagslen != datafieldslen {
		return errors.New("Error validating event - TypeFlags and Data Fields lengths do not match")
	}
	if typeflagslen > 10 {
		return errors.New("Error validating event - Maximum TypeFlags count is 10 but found: " + strconv.Itoa(typeflagslen))
	}
	return nil
}

// ValidateReadMessageChoiceTriggerEvent function
func (h *EventParser) ValidateReadMessageChoiceTriggerEvent(event Event) (err error) {
	typeflagslen := len(event.TypeFlags)
	datafieldslen := len(event.Data)
	if len(event.TypeFlags) < 1 {
		return errors.New("Error validating event - Expected at least 1 typeflag")
	}
	if typeflagslen != datafieldslen {
		return errors.New("Error validating event - TypeFlags and Data Fields lengths do not match")
	}
	if typeflagslen > 10 {
		return errors.New("Error validating event - Maximum TypeFlags count is 10 but found: " + strconv.Itoa(typeflagslen))
	}
	if len(event.Data) < 1 {
		return errors.New("error validating event - Expected at least one data field")
	}
	// Now check to see that either the supplied eventID's are valid or set to nil
	for _, field := range event.Data {
		if field != "nil" {
			if !h.eventsdb.ValidateEventByID(field) {
				return errors.New("Error validating event - Invalid event found in data: " + field)
			}
		}
	}
	return nil
}

// ValidateSendMessage function
func (h *EventParser) ValidateSendMessage(event Event) (err error) {
	if len(event.Data) < 1 {
		return errors.New("error validating event - Expected a data field")
	}
	return nil
}

// ValidateTimedSendMessageEvent function
func (h *EventParser) ValidateTimedSendMessageEvent(event Event) (err error) {
	if len(event.TypeFlags) < 1 {
		return errors.New("error validating event - Expected at least one type flag")
	}
	timeout, err := strconv.Atoi(event.TypeFlags[1])
	if err != nil {
		return errors.New("error validating event - Could not parse timeout: " + err.Error())
	}
	if timeout > 300 {
		return errors.New("error validating event - Maximum timeout is 300 but found: " + strconv.Itoa(timeout))
	}
	return nil
}

// ValidateMessageTriggerSuccessFail function
func (h *EventParser) ValidateMessageTriggerSuccessFail(event Event) (err error) {
	if len(event.TypeFlags) < 1 {
		return errors.New("error validating event - expected one type flag")
	}
	return nil
}

// ValidateTriggerSuccess function
func (h *EventParser) ValidateTriggerSuccess(event Event) (err error) {
	if len(event.TypeFlags) > 0 {
		return errors.New("error validating event - expected no type flags")
	}
	if len(event.Data) > 0 {
		return errors.New("error validating event - expected no data fields")
	}
	return nil
}

// ValidateTriggerFailure function
func (h *EventParser) ValidateTriggerFailure(event Event) (err error) {
	if len(event.TypeFlags) > 0 {
		return errors.New("error validating event - expected no type flags")
	}
	if len(event.Data) > 0 {
		return errors.New("error validating event - expected no data fields")
	}
	return nil
}

// ValidateSendMessageTriggerEvent function
func (h *EventParser) ValidateSendMessageTriggerEvent(event Event) (err error) {
	if len(event.TypeFlags) < 1 {
		return errors.New("Error validating event - Expected a type flag")
	}
	if len(event.Data) < 1 {
		return errors.New("error validating event - Expected a data field")
	}
	// Now check to see that either the supplied eventID's are valid or set to nil
	for _, field := range event.Data {
		if field != "nil" {
			if !h.eventsdb.ValidateEventByID(field) {
				return errors.New("Error validating event - Invalid event found in data: " + field)
			}
		}
	}
	return nil
}

// ValidateTriggerFailureSendError function
func (h *EventParser) ValidateTriggerFailureSendError(event Event) (err error) {
	if len(event.Data) < 1 {
		return errors.New("error validating event - Expected a data field")
	}
	return nil
}

// ValidateMessageChoiceDefault function
func (h *EventParser) ValidateMessageChoiceDefault(event Event) (err error) {
	typeflagslen := len(event.TypeFlags)
	datafieldslen := len(event.Data)
	if len(event.TypeFlags) < 1 {
		return errors.New("Error validating event - Expected at least 1 typeflag")
	}
	if typeflagslen != datafieldslen {
		return errors.New("Error validating event - TypeFlags and Data Fields lengths do not match")
	}
	if typeflagslen > 10 {
		return errors.New("Error validating event - Maximum TypeFlags count is 10 but found: " + strconv.Itoa(typeflagslen))
	}
	if event.DefaultData == "" {
		return errors.New("MessageChoiceDefault requires a default message in DefaultData")
	}
	return nil
}

// ValidateMessageChoiceDefaultEvent function
func (h *EventParser) ValidateMessageChoiceDefaultEvent(event Event) (err error) {
	typeflagslen := len(event.TypeFlags)
	datafieldslen := len(event.Data)
	if len(event.TypeFlags) < 1 {
		return errors.New("Error validating event - Expected at least 1 typeflag")
	}
	if typeflagslen != datafieldslen {
		return errors.New("Error validating event - TypeFlags and Data Fields lengths do not match")
	}
	if typeflagslen > 10 {
		return errors.New("Error validating event - Maximum TypeFlags count is 10 but found: " + strconv.Itoa(typeflagslen))
	}
	if len(event.Data) < 1 {
		return errors.New("error validating event - Expected at least one data field")
	}
	// Now check to see that either the supplied eventID's are valid or set to nil
	for _, field := range event.Data {
		if field != "nil" {
			if !h.eventsdb.ValidateEventByID(field) {
				return errors.New("Error validating event - Invalid event found in data: " + field)
			}
		}
	}
	if event.DefaultData == "" {
		return errors.New("MessageChoiceDefault requires a default event in DefaultData")
	}
	if event.DefaultData != "nil" {
		if !h.eventsdb.ValidateEventByID(event.DefaultData) {
			return errors.New("Error validating event - Invalid event found in DefaultData: " + event.DefaultData)
		}
	}
	return nil
}
