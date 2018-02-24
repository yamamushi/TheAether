package main

import (
	"encoding/json"
	"strings"
)

// Parse event scripts and formatted event data fields
type EventParser struct {


}


func (h *EventParser) ParseFormattedEvent(data string, channelID string, userID string) (parsed Event, err error){
	unmarshallcontainer := Event{}
	if err := json.Unmarshal([]byte(data), &unmarshallcontainer); err != nil {
		return unmarshallcontainer, err
	}

	unmarshallcontainer.ChannelID = channelID
	unmarshallcontainer.CreatorID = userID
	unmarshallcontainer.Data = strings.Replace(unmarshallcontainer.Data, "_user_", "<@"+userID+">", -1)

	id := strings.Split(GetUUIDv2(), "-")
	unmarshallcontainer.ID = id[0]

	return unmarshallcontainer, nil
}