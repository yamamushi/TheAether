package main

import (
	"github.com/bwmarrin/discordgo"
)

// LaunchChildEvent function
func (h *EventHandler) LaunchChildEvent(parenteventID string, childeventID string, eventmessagesid string, channelID string, s *discordgo.Session, m *discordgo.MessageCreate) {
	// First we load the keyed eventID in the data array
	if childeventID != "nil" {
		triggeredevent, err := h.eventsdb.GetEventByID(childeventID)
		if err != nil {
			s.ChannelMessageSend(channelID, "Error triggering keyed event - Source: "+
				parenteventID+" Trigger: "+childeventID+" Error: "+err.Error())
			return
		}

		if triggeredevent.Type == "ReadMessage" {
			h.WatchEvent(h.UnfoldReadMessage, eventmessagesid, triggeredevent.ID, channelID)
		} else if triggeredevent.Type == "ReadTimedMessage" {
			h.WatchEvent(h.UnfoldReadTimedMessage, eventmessagesid, triggeredevent.ID, channelID)
		} else if triggeredevent.Type == "ReadMessageChoice" {
			h.WatchEvent(h.UnfoldReadMessageChoiceTriggerMessage, eventmessagesid, triggeredevent.ID, channelID)
		} else if triggeredevent.Type == "ReadMessageChoiceTriggerEvent" {
			h.WatchEvent(h.UnfoldReadMessageChoiceTriggerEvent, eventmessagesid, triggeredevent.ID, channelID)
		} else if triggeredevent.Type == "SendMessage" {
			h.UnfoldSendMessage(triggeredevent.ID, eventmessagesid, s, m)
		} else if triggeredevent.Type == "TimedSendMessage" {
			h.UnfoldTimedSendMessage(triggeredevent.ID, eventmessagesid, s, m)
		} else if triggeredevent.Type == "ReadMessageTriggerSuccessFail" {
			h.WatchEvent(h.UnfoldReadMessageTriggerSuccessFail, eventmessagesid, triggeredevent.ID, channelID)
		} else if triggeredevent.Type == "TriggerSuccess" {
			h.UnfoldTriggerSuccess(triggeredevent.ID, eventmessagesid, s, m)
		} else if triggeredevent.Type == "TriggerFailure" {
			h.UnfoldTriggerFailure(triggeredevent.ID, eventmessagesid, s, m)
		} else if triggeredevent.Type == "SendMessageTriggerEvent" {
			h.UnfoldSendMessageTriggerEvent(triggeredevent.ID, eventmessagesid, s, m)
		} else if triggeredevent.Type == "TriggerFailureSendError" {
			h.UnfoldTriggerFailureSendError(triggeredevent.ID, eventmessagesid, s, m)
		} else if triggeredevent.Type == "MessageChoiceDefaultEvent" {
			h.WatchEvent(h.UnfoldMessageChoiceDefaultEvent, eventmessagesid, triggeredevent.ID, channelID)
		} else if triggeredevent.Type == "MessageChoiceDefault" {
			h.WatchEvent(h.UnfoldMessageChoiceDefault, eventmessagesid, triggeredevent.ID, channelID)
		}
	}
}
