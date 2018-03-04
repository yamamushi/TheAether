package main

import (
	"github.com/bwmarrin/discordgo"
	"reflect"
)

// EventCallback struct
type EventCallback struct {
	ChannelID       string
	EventID         string
	EventMessagesID string
	Handler         func(string, string, *discordgo.Session, *discordgo.MessageCreate)
}

// These functions are responsible for working with the callback system for events
// It works similar to the callback_handler system, except that this also has the event messages system
// Which is how events communicate with each other.

// WatchEvent function
func (h *EventHandler) WatchEvent(Handler func(string, string, *discordgo.Session, *discordgo.MessageCreate), EventMessagesID string, EventID string, ChannelID string) {
	// Make sure we don't duplicate events in the watch list
	for e := h.WatchList.Front(); e != nil; e = e.Next() {
		r := reflect.ValueOf(e.Value)
		eventmessagesid := reflect.Indirect(r).FieldByName("EventMessagesID")
		eventID := reflect.Indirect(r).FieldByName("EventID")
		channel := reflect.Indirect(r).FieldByName("ChannelID")
		if channel.String() == ChannelID && eventID.String() == EventID && eventmessagesid.String() == EventMessagesID {
			return
		}
	}
	item := EventCallback{ChannelID: ChannelID, EventID: EventID, Handler: Handler, EventMessagesID: EventMessagesID}
	h.WatchList.PushBack(item)
}

// UnWatchEvent function
func (h *EventHandler) UnWatchEvent(ChannelID string, EventID string, EventMessagesID string) {
	ChannelID = CleanChannel(ChannelID)
	// Clear usermanager element by iterating
	for e := h.WatchList.Front(); e != nil; e = e.Next() {
		r := reflect.ValueOf(e.Value)
		eventmessagesid := reflect.Indirect(r).FieldByName("EventMessagesID")
		channel := reflect.Indirect(r).FieldByName("ChannelID")
		eventID := reflect.Indirect(r).FieldByName("EventID")

		if ChannelID == "" || EventMessagesID == "" {
			if eventID.String() == EventID {
				h.WatchList.Remove(e)
				h.RemoveEventFromRoom(EventID, ChannelID)
			}
		} else {
			if channel.String() == ChannelID && eventID.String() == EventID && eventmessagesid.String() == EventMessagesID {
				h.WatchList.Remove(e)
				h.RemoveEventFromRoom(EventID, ChannelID)
			}
		}
	}
}

// ListEnabled function
func (h *EventHandler) ListEnabled(channelID string) (formatted string, err error) {
	formatted = "```\n"
	for e := h.WatchList.Front(); e != nil; e = e.Next() {
		r := reflect.ValueOf(e.Value)
		eventmessagesid := reflect.Indirect(r).FieldByName("EventMessagesID")
		eventID := reflect.Indirect(r).FieldByName("EventID")
		channel := reflect.Indirect(r).FieldByName("ChannelID")
		if channel.String() == channelID {
			formatted = formatted + "ID: " + eventID.String() + " EventMessagesID: " + eventmessagesid.String() + "\n"
		}
	}
	formatted = formatted + "\n```\n"
	return formatted, nil
}

// ReadEvents function
func (h *EventHandler) ReadEvents(s *discordgo.Session, m *discordgo.MessageCreate) {
	for e := h.WatchList.Front(); e != nil; e = e.Next() {
		r := reflect.ValueOf(e.Value)
		channelid := reflect.Indirect(r).FieldByName("ChannelID")

		if m.ChannelID == channelid.String() {
			// We get the handler interface from our "Handler" field
			handler := reflect.Indirect(r).FieldByName("Handler")

			// We get our argument list from the Args field
			arglist := reflect.Indirect(r).FieldByName("EventID")
			eventid := arglist.String()

			eventmessagesid := reflect.Indirect(r).FieldByName("EventMessagesID")
			eventmessageid := eventmessagesid.String()

			// We now type the interface to the handler type
			//v := reflect.ValueOf(handler)
			rargs := make([]reflect.Value, 4)

			//var sizeofargs = len(rargs)
			rargs[0] = reflect.ValueOf(eventid)
			rargs[1] = reflect.ValueOf(eventmessageid)
			rargs[2] = reflect.ValueOf(s)
			rargs[3] = reflect.ValueOf(m)

			go handler.Call(rargs)
			//handlerid := reflect.Indirect(r).FieldByName("HandlerID").String()
			//c.UnWatchEvent(m.ChannelID, handlerid)
		}
	}
}
