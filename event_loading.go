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
		switch triggeredevent.Type {
		case "ReadMessage":
			h.WatchEvent(h.UnfoldReadMessage, eventmessagesid, triggeredevent.ID, channelID)
		case "ReadTimedMessage":
			h.WatchEvent(h.UnfoldReadTimedMessage, eventmessagesid, triggeredevent.ID, channelID)
		case "ReadMessageChoice":
			h.WatchEvent(h.UnfoldReadMessageChoiceTriggerMessage, eventmessagesid, triggeredevent.ID, channelID)
		case "ReadMessageChoiceTriggerEvent":
			h.WatchEvent(h.UnfoldReadMessageChoiceTriggerEvent, eventmessagesid, triggeredevent.ID, channelID)
		case "SendMessage":
			h.UnfoldSendMessage(triggeredevent.ID, eventmessagesid, s, m)
		case "TimedSendMessage":
			h.UnfoldTimedSendMessage(triggeredevent.ID, eventmessagesid, s, m)
		case "ReadMessageTriggerSuccessFail":
			h.WatchEvent(h.UnfoldReadMessageTriggerSuccessFail, eventmessagesid, triggeredevent.ID, channelID)
		case "TriggerSuccess":
			h.UnfoldTriggerSuccess(triggeredevent.ID, eventmessagesid, s, m)
		case "TriggerFailure":
			h.UnfoldTriggerFailure(triggeredevent.ID, eventmessagesid, s, m)
		case "SendMessageTriggerEvent":
			h.UnfoldSendMessageTriggerEvent(triggeredevent.ID, eventmessagesid, s, m)
		case "TriggerFailureSendError":
			h.UnfoldTriggerFailureSendError(triggeredevent.ID, eventmessagesid, s, m)
		case "MessageChoiceDefault":
			h.WatchEvent(h.UnfoldMessageChoiceDefault, eventmessagesid, triggeredevent.ID, channelID)
		case "MessageChoiceDefaultEvent":
			h.WatchEvent(h.UnfoldMessageChoiceDefaultEvent, eventmessagesid, triggeredevent.ID, channelID)
		case "RollDiceSum":
			h.UnfoldRollDiceSum(triggeredevent.ID, eventmessagesid, s, m)
		default:
			return
		}
	}
}
