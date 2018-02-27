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
}

// ParseFormattedEvent function
func (h *EventParser) ParseFormattedEvent(data string, channelID string, userID string) (parsed Event, err error) {
	unmarshallcontainer := Event{}
	if err := json.Unmarshal([]byte(data), &unmarshallcontainer); err != nil {
		return unmarshallcontainer, err
	}

	unmarshallcontainer.CreatorID = userID
	unmarshallcontainer.RunCount = 0
	unmarshallcontainer.ChannelID = channelID

	//unmarshallcontainer.Data = strings.Replace(unmarshallcontainer.Data, "_user_", "<@"+userID+">", -1)
	// We will fix this in another parser later, this should not be formatted in this function

	// Generate and assign an ID to this event
	id := strings.Split(GetUUIDv2(), "-")
	unmarshallcontainer.ID = id[0]

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
		if len(event.TypeFlags) != 1 {
			return errors.New("Error validating event - Expected 1 typeflag but found: " + strconv.Itoa(len(event.TypeFlags)))
		}
		return nil
	} else if event.Type == "TimedMessage" {
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
	} else if event.Type == "ReadMessageChoice" {
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

	return nil
}
