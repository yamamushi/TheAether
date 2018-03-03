package main

import (
	"github.com/bwmarrin/discordgo"
	"strconv"
	"strings"
	"time"
)

// UnfoldReadMessage function
func (h *EventHandler) UnfoldReadMessage(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}
	// Ignore bots
	if m.Author.Bot {
		return
	}

	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event: "+eventID+" Error: "+err.Error())
		return
	}

	keyword := event.TypeFlags[0]
	messageContent := strings.Fields(strings.ToLower(m.Content))

	for _, messagefield := range messageContent {
		// We don't need to check for the userID here because that's what checking for event.Attachable did
		if messagefield == keyword {
			// First we send the data
			s.ChannelMessageSend(m.ChannelID, FormatEventMessage(event.Data[0], m.Author.ID, m.ChannelID))

			//h.CheckCycles(event, eventmessagesid, s, m)
			h.UnWatchEvent(m.ChannelID, event.ID, eventmessagesid)
			_ = h.DisableEvent(event.ID, m.ChannelID)
			/*if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error disabling event: "+event.ID+" Error: "+err.Error())
				return
			}*/
			_ = h.eventsdb.SaveEventToDB(event)
			/*if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error saving event: "+event.ID+" Error: "+err.Error())
				return
			}*/
			// We are at the end so we terminate
			h.eventmessages.TerminateEvents(eventmessagesid)
			return
		}
	}
	return
}

// UnfoldReadTimedMessage function
func (h *EventHandler) UnfoldReadTimedMessage(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}
	// Ignore bots
	if m.Author.Bot {
		return
	}

	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		return
	}

	keyword := event.TypeFlags[0]
	timeout, _ := strconv.Atoi(event.TypeFlags[1]) // We don't bother checking for an error here because that was handled during the event registration.
	messageContent := strings.Fields(strings.ToLower(m.Content))

	for _, messagefield := range messageContent {
		if messagefield == keyword {
			// First we want to sleep for our timeout period
			time.Sleep(time.Duration(timeout) * time.Second)
			// Now we send the data
			s.ChannelMessageSend(m.ChannelID, FormatEventMessage(event.Data[0], m.Author.ID, m.ChannelID))

			h.DisableEvent(event.ID, m.ChannelID)
			h.UnWatchEvent(m.ChannelID, event.ID, eventmessagesid)
			h.eventsdb.SaveEventToDB(event)
			h.eventmessages.TerminateEvents(eventmessagesid)

			return
		}
	}
	return
}

// UnfoldSendMessage function
func (h *EventHandler) UnfoldSendMessage(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}
	// Ignore bots
	if m.Author.Bot {
		return
	}

	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		return
	}

	// Now we send the data
	s.ChannelMessageSend(m.ChannelID, FormatEventMessage(event.Data[0], m.Author.ID, m.ChannelID))

	h.DisableEvent(event.ID, m.ChannelID)
	h.UnWatchEvent(m.ChannelID, event.ID, eventmessagesid)
	h.eventsdb.SaveEventToDB(event)
	h.eventmessages.TerminateEvents(eventmessagesid)

	return
}

// UnfoldTimedSendMessage function
func (h *EventHandler) UnfoldTimedSendMessage(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}
	// Ignore bots
	if m.Author.Bot {
		return
	}

	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		return
	}

	timeout, _ := strconv.Atoi(event.TypeFlags[1]) // We don't bother checking for an error here because that was handled during the event registration.

	// First we want to sleep for our timeout period
	time.Sleep(time.Duration(timeout) * time.Second)
	// Now we send the data
	s.ChannelMessageSend(m.ChannelID, FormatEventMessage(event.Data[0], m.Author.ID, m.ChannelID))

	h.DisableEvent(event.ID, m.ChannelID)
	h.UnWatchEvent(m.ChannelID, event.ID, eventmessagesid)
	h.eventsdb.SaveEventToDB(event)
	h.eventmessages.TerminateEvents(eventmessagesid)

	return
}

// UnfoldReadMessageChoiceTriggerMessage function
func (h *EventHandler) UnfoldReadMessageChoiceTriggerMessage(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}
	// Ignore bots
	if m.Author.Bot {
		return
	}

	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		return
	}

	messageContent := strings.Fields(strings.ToLower(m.Content))

	for i, field := range event.TypeFlags {
		for _, message := range messageContent {
			if field == message {
				// First we send the data that is keyed to the field
				s.ChannelMessageSend(m.ChannelID, FormatEventMessage(event.Data[i], m.Author.ID, m.ChannelID))

				h.DisableEvent(event.ID, m.ChannelID)
				h.UnWatchEvent(m.ChannelID, event.ID, eventmessagesid)
				h.eventsdb.SaveEventToDB(event)
				h.eventmessages.TerminateEvents(eventmessagesid)

				return
			}
		}
	}
}

// UnfoldReadMessageChoiceTriggerEvent function
func (h *EventHandler) UnfoldReadMessageChoiceTriggerEvent(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if !h.IsValidEventMessage(s, m) {
		return
	}

	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		return
	}

	messageContent := strings.Fields(strings.ToLower(m.Content))

	for i, field := range event.TypeFlags {
		for _, message := range messageContent {
			if field == message {
				// First we load the keyed eventID in the data array
				if event.Data[i] != "nil" {
					go h.LaunchChildEvent(event.ID, event.Data[i], eventmessagesid, s, m)
				} else {
					h.DisableEvent(event.ID, m.ChannelID)
					h.UnWatchEvent(m.ChannelID, event.ID, eventmessagesid)
					h.eventsdb.SaveEventToDB(event)
					h.eventmessages.TerminateEvents(eventmessagesid)
				}
				return
			}
		}
	}
}

// UnfoldReadMessageTriggerSuccessFail function
func (h *EventHandler) UnfoldReadMessageTriggerSuccessFail(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if !h.IsValidEventMessage(s, m) {
		return
	}

	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		return
	}

	messageContent := strings.Fields(strings.ToLower(m.Content))

	for _, field := range event.TypeFlags {
		for _, message := range messageContent {
			if field == message {
				h.eventmessages.SetSuccessfulStatus(eventmessagesid)

				h.DisableEvent(event.ID, m.ChannelID)
				h.UnWatchEvent(m.ChannelID, event.ID, eventmessagesid)
				h.eventsdb.SaveEventToDB(event)
				h.eventmessages.TerminateEvents(eventmessagesid)

				return
			}
		}
	}
}

// UnfoldTriggerSuccess function
func (h *EventHandler) UnfoldTriggerSuccess(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if !h.IsValidEventMessage(s, m) {
		return
	}

	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		return
	}

	h.eventmessages.SetSuccessfulStatus(eventmessagesid)

	h.DisableEvent(event.ID, m.ChannelID)
	h.UnWatchEvent(m.ChannelID, event.ID, eventmessagesid)
	h.eventsdb.SaveEventToDB(event)
	h.eventmessages.TerminateEvents(eventmessagesid)

	return

}

// UnfoldTriggerFailure function
func (h *EventHandler) UnfoldTriggerFailure(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if !h.IsValidEventMessage(s, m) {
		return
	}

	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		return
	}

	h.eventmessages.SetFailureStatus(eventmessagesid)

	h.DisableEvent(event.ID, m.ChannelID)
	h.UnWatchEvent(m.ChannelID, event.ID, eventmessagesid)
	h.eventsdb.SaveEventToDB(event)
	h.eventmessages.TerminateEvents(eventmessagesid)
	return

}

// UnfoldTriggerFailureSendError function
func (h *EventHandler) UnfoldTriggerFailureSendError(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if !h.IsValidEventMessage(s, m) {
		return
	}

	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		return
	}

	h.eventmessages.SetErrorMessage(eventmessagesid, event.Data[0])
	h.eventmessages.SetFailureStatus(eventmessagesid)

	h.DisableEvent(event.ID, m.ChannelID)
	h.UnWatchEvent(m.ChannelID, event.ID, eventmessagesid)
	h.eventsdb.SaveEventToDB(event)
	h.eventmessages.TerminateEvents(eventmessagesid)
	return

}

// UnfoldSendMessageTriggerEvent function
func (h *EventHandler) UnfoldSendMessageTriggerEvent(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}
	// Ignore bots
	if m.Author.Bot {
		return
	}

	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		return
	}

	// Now we send the message
	s.ChannelMessageSend(m.ChannelID, FormatEventMessage(event.Data[0], m.Author.ID, m.ChannelID))
	h.DisableEvent(event.ID, m.ChannelID)
	h.UnWatchEvent(m.ChannelID, event.ID, eventmessagesid)
	h.eventsdb.SaveEventToDB(event)
	if event.Data[0] != "nil" {
		go h.LaunchChildEvent(event.ID, event.Data[0], eventmessagesid, s, m)
	} else {

		h.eventmessages.TerminateEvents(eventmessagesid)
	}

	return
}

// UnfoldMessageChoiceDefaultEvent function
func (h *EventHandler) UnfoldMessageChoiceDefaultEvent(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {
	if !h.IsValidEventMessage(s, m) {
		return
	}

	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		return
	}

	messageContent := strings.Fields(strings.ToLower(m.Content))

	h.DisableEvent(event.ID, m.ChannelID)
	h.UnWatchEvent(m.ChannelID, event.ID, eventmessagesid)
	h.eventsdb.SaveEventToDB(event)

	for i, field := range event.TypeFlags {
		for _, message := range messageContent {
			if field == message {
				// First we load the keyed eventID in the data array
				if event.Data[i] != "nil" {
					go h.LaunchChildEvent(event.ID, event.Data[i], eventmessagesid, s, m)
				} else {
					if event.DefaultData != "nil" {
						go h.LaunchChildEvent(event.ID, event.DefaultData, eventmessagesid, s, m)
					}
				}
				return
			}
		}
	}
}

// UnfoldMessageChoiceDefault function
func (h *EventHandler) UnfoldMessageChoiceDefault(eventID string, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}
	// Ignore bots
	if m.Author.Bot {
		return
	}

	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		return
	}

	messageContent := strings.Fields(strings.ToLower(m.Content))

	for i, field := range event.TypeFlags {
		for _, message := range messageContent {
			if field == message {
				// First we send the data that is keyed to the field
				s.ChannelMessageSend(m.ChannelID, FormatEventMessage(event.Data[i], m.Author.ID, m.ChannelID))

				h.DisableEvent(event.ID, m.ChannelID)
				h.UnWatchEvent(m.ChannelID, event.ID, eventmessagesid)
				h.eventsdb.SaveEventToDB(event)
				h.eventmessages.TerminateEvents(eventmessagesid)

				return
			}
		}
	}
	s.ChannelMessageSend(m.ChannelID, FormatEventMessage(event.DefaultData, m.Author.ID, m.ChannelID))
	h.DisableEvent(event.ID, m.ChannelID)
	h.UnWatchEvent(m.ChannelID, event.ID, eventmessagesid)
	h.eventsdb.SaveEventToDB(event)
	h.eventmessages.TerminateEvents(eventmessagesid)
	return
}
