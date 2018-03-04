package main

import (
	"github.com/bwmarrin/discordgo"
	"strconv"
	"strings"
	"time"
	//"fmt"
)

// UnfoldReadMessage function
func (h *EventHandler) UnfoldReadMessage(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event: "+eventID+" Error: "+err.Error())
		return
	}

	for _, field := range event.TypeFlags {
		// We don't need to check for the userID here because that's what checking for event.Attachable did
		if strings.Contains(m.Content, field) {
			// First we send the data
			s.ChannelMessageSend(m.ChannelID, FormatEventMessage(event.Data[0], m.Author.ID, m.ChannelID))
			h.StopEvents(event, m.ChannelID, eventmessagesid)
			return
		}
	}
	h.StopEvents(event, m.ChannelID, eventmessagesid)
	return
}

// UnfoldReadTimedMessage function
func (h *EventHandler) UnfoldReadTimedMessage(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		h.StopEvents(event, m.ChannelID, eventmessagesid)
		return
	}

	keyword := event.TypeFlags[0]
	timeout, _ := strconv.Atoi(event.TypeFlags[1]) // We don't bother checking for an error here because that was handled during the event registration.
	messageContent := strings.Fields(strings.ToLower(m.Content))

	for _, messagefield := range messageContent {
		if strings.Contains(messagefield, keyword) {
			// First we want to sleep for our timeout period
			time.Sleep(time.Duration(timeout) * time.Second)
			// Now we send the data
			s.ChannelMessageSend(m.ChannelID, FormatEventMessage(event.Data[0], m.Author.ID, m.ChannelID))
			h.StopEvents(event, m.ChannelID, eventmessagesid)
			return
		}
	}
	h.StopEvents(event, m.ChannelID, eventmessagesid)
	return
}

// UnfoldSendMessage function
func (h *EventHandler) UnfoldSendMessage(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		h.StopEvents(event, m.ChannelID, eventmessagesid)
		return
	}

	// Now we send the data
	s.ChannelMessageSend(m.ChannelID, FormatEventMessage(event.Data[0], m.Author.ID, m.ChannelID))
	h.StopEvents(event, m.ChannelID, eventmessagesid)
	return
}

// UnfoldTimedSendMessage function
func (h *EventHandler) UnfoldTimedSendMessage(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		h.StopEvents(event, m.ChannelID, eventmessagesid)
		return
	}

	timeout, _ := strconv.Atoi(event.TypeFlags[1]) // We don't bother checking for an error here because that was handled during the event registration.

	// First we want to sleep for our timeout period
	time.Sleep(time.Duration(timeout) * time.Second)
	// Now we send the data
	s.ChannelMessageSend(m.ChannelID, FormatEventMessage(event.Data[0], m.Author.ID, m.ChannelID))
	h.StopEvents(event, m.ChannelID, eventmessagesid)
	return
}

// UnfoldReadMessageChoiceTriggerMessage function
func (h *EventHandler) UnfoldReadMessageChoiceTriggerMessage(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		h.StopEvents(event, m.ChannelID, eventmessagesid)
		return
	}

	//messageContent := strings.Fields(strings.ToLower(m.Content))

	for i, field := range event.TypeFlags {
		if strings.Contains(m.Content, field) {
			// First we send the data that is keyed to the field
			s.ChannelMessageSend(m.ChannelID, FormatEventMessage(event.Data[i], m.Author.ID, m.ChannelID))
			h.StopEvents(event, m.ChannelID, eventmessagesid)
			return
		}
	}
	h.StopEvents(event, m.ChannelID, eventmessagesid)
	return
}

// UnfoldReadMessageChoiceTriggerEvent function
func (h *EventHandler) UnfoldReadMessageChoiceTriggerEvent(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		h.StopEvents(event, m.ChannelID, eventmessagesid)
		return
	}

	for i, field := range event.TypeFlags {
		if strings.Contains(m.Content, field) {
			// First we load the keyed eventID in the data array
			if event.Data[i] != "nil" {
				go h.LaunchChildEvent(event.ID, event.Data[i], eventmessagesid, m.ChannelID, s, m)
				h.EventComplete(event, m.ChannelID, eventmessagesid)
			} else {
				h.StopEvents(event, m.ChannelID, eventmessagesid)
			}
			return
		}
	}
	h.StopEvents(event, m.ChannelID, eventmessagesid)
	return
}

// UnfoldReadMessageTriggerSuccessFail function
func (h *EventHandler) UnfoldReadMessageTriggerSuccessFail(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		return
	}

	for _, field := range event.TypeFlags {
		if strings.Contains(m.Content, field) {
			h.eventmessages.SetSuccessfulStatus(eventmessagesid)
			h.StopEvents(event, m.ChannelID, eventmessagesid)
			return
		}
	}
	h.eventmessages.SetFailureStatus(eventmessagesid)
	h.StopEvents(event, m.ChannelID, eventmessagesid)
	return
}

// UnfoldTriggerSuccess function
func (h *EventHandler) UnfoldTriggerSuccess(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		return
	}

	h.eventmessages.SetSuccessfulStatus(eventmessagesid)

	h.StopEvents(event, m.ChannelID, eventmessagesid)
	return
}

// UnfoldTriggerFailure function
func (h *EventHandler) UnfoldTriggerFailure(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		return
	}

	h.eventmessages.SetFailureStatus(eventmessagesid)
	h.StopEvents(event, m.ChannelID, eventmessagesid)
	return
}

// UnfoldTriggerFailureSendError function
func (h *EventHandler) UnfoldTriggerFailureSendError(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		return
	}

	h.eventmessages.SetErrorMessage(eventmessagesid, event.Data[0])
	h.eventmessages.SetFailureStatus(eventmessagesid)
	h.StopEvents(event, m.ChannelID, eventmessagesid)
	return
}

// UnfoldSendMessageTriggerEvent function
func (h *EventHandler) UnfoldSendMessageTriggerEvent(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		return
	}

	// Now we send the message
	s.ChannelMessageSend(m.ChannelID, FormatEventMessage(event.TypeFlags[0], m.Author.ID, m.ChannelID))
	h.EventComplete(event, m.ChannelID, eventmessagesid)
	if event.Data[0] != "nil" {
		go h.LaunchChildEvent(event.ID, event.Data[0], eventmessagesid, m.ChannelID, s, m)
		h.EventComplete(event, m.ChannelID, eventmessagesid)
	} else {
		h.StopEvents(event, m.ChannelID, eventmessagesid)
	}
	return
}

// UnfoldMessageChoiceDefaultEvent function
func (h *EventHandler) UnfoldMessageChoiceDefaultEvent(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		h.StopEvents(event, m.ChannelID, eventmessagesid)
		return
	}

	for i, field := range event.TypeFlags {
		if strings.Contains(m.Content, field) {
			//fmt.Println("Field: " + field + " Message: " + message)
			// First we load the keyed eventID in the data array
			if event.Data[i] != "nil" {
				go h.LaunchChildEvent(event.ID, event.Data[i], eventmessagesid, m.ChannelID, s, m)
			} else {
				go h.LaunchChildEvent(event.ID, event.DefaultData, eventmessagesid, m.ChannelID, s, m)
			}
			h.EventComplete(event, m.ChannelID, eventmessagesid)
			return
		}
	}
	if event.DefaultData != "nil" {
		//fmt.Print("Launching default event: " + event.DefaultData)
		go h.LaunchChildEvent(event.ID, event.DefaultData, eventmessagesid, m.ChannelID, s, m)
	}
	h.StopEvents(event, m.ChannelID, eventmessagesid)
	return
}

// UnfoldMessageChoiceDefault function
func (h *EventHandler) UnfoldMessageChoiceDefault(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		h.StopEvents(event, m.ChannelID, eventmessagesid)
		return
	}

	for i, field := range event.TypeFlags {
		if strings.Contains(m.Content, field) {
			// First we send the data that is keyed to the field
			s.ChannelMessageSend(m.ChannelID, FormatEventMessage(event.Data[i], m.Author.ID, m.ChannelID))
			h.StopEvents(event, m.ChannelID, eventmessagesid)
			return
		}
	}
	s.ChannelMessageSend(m.ChannelID, FormatEventMessage(event.DefaultData, m.Author.ID, m.ChannelID))
	h.StopEvents(event, m.ChannelID, eventmessagesid)
	return
}

// UnfoldRollDiceSum function
func (h *EventHandler) UnfoldRollDiceSum(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		return
	}

	faces, err := strconv.Atoi(event.TypeFlags[0])
	if err != nil {
		h.StopEvents(event, m.ChannelID, eventmessagesid)
		return
	}
	count, err := strconv.Atoi(event.TypeFlags[1])
	if err != nil {
		h.StopEvents(event, m.ChannelID, eventmessagesid)
		return
	}

	roll := RollDiceAndAdd(faces, count)
	err = h.eventmessages.SetDieRoll(eventmessagesid, roll)
	if err != nil {
		h.StopEvents(event, m.ChannelID, eventmessagesid)
		return
	}

	if event.Data[0] != "nil" {
		go h.LaunchChildEvent(event.ID, event.Data[0], eventmessagesid, m.ChannelID, s, m)
		h.EventComplete(event, m.ChannelID, eventmessagesid)
		return
	}
	h.StopEvents(event, m.ChannelID, eventmessagesid)
	return
}
