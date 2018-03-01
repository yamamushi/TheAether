package main

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// This handles the discord input interface for registering events and managing them

// EventHandler struct
type EventHandler struct {
	conf     *Config
	registry *CommandRegistry
	callback *CallbackHandler
	db       *DBHandler
	user     *UserHandler

	WatchList list.List
	dg        *discordgo.Session
	logger    *Logger

	eventsdb *EventsDB
	parser   *EventParser
}

// EventCallback struct
type EventCallback struct {
	ChannelID  string
	EventID    string
	KeyValueID string
	Handler    func(string, string, *discordgo.Session, *discordgo.MessageCreate)
}

// Init function
func (h *EventHandler) Init() (err error) {
	fmt.Println("Registering Event Handler Command")
	h.eventsdb = new(EventsDB)
	h.eventsdb.db = h.db
	h.parser = new(EventParser)
	h.parser.eventsdb = h.eventsdb
	h.RegisterCommand()

	fmt.Println("Loading Registered Events from Database")
	err = h.LoadEventsAtBoot()
	if err != nil {
		return err
	}
	return nil
}

// RegisterCommand command function
func (h *EventHandler) RegisterCommand() {
	h.registry.Register("events", "Manage events", "add|remove|list|info|enabled|disable|listenabled")
	h.registry.AddGroup("events", "builder")
}

// Read function
func (h *EventHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {
	cp := h.conf.MainConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding usermanager")
		return
	}

	if strings.HasPrefix(m.Content, cp+"events") {
		if h.registry.CheckPermission("events", user, s, m) {

			command := strings.Fields(m.Content)

			// Grab our sender ID to verify if this usermanager has permission to use this command
			db := h.db.rawdb.From("Users")
			var user User
			err := db.One("ID", m.Author.ID, &user)
			if err != nil {
				fmt.Println("error retrieving usermanager:" + m.Author.ID)
			}

			if user.CheckRole("builder") {
				h.ParseCommand(command, s, m)
			}
		}
	}
}

// ParseCommand function
func (h *EventHandler) ParseCommand(input []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	argument, payload := GetArgumentAndFlags(input)

	if argument == "add" {
		if len(payload) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Command 'add' expects an argument")
			return
		}
		// We pass in the full message here because we intend to unpack it later
		eventID, err := h.RegisterEvent(m.Content, s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error registering new event: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Event registered with ID: "+eventID)
		return
	}
	if argument == "remove" {
		if len(payload) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Command 'remove' expects an argument")
			return
		}
		err := h.RemoveEvent(payload[0], m.Author.ID, s, m.ChannelID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error removing event: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Event record removed!")
		return
	}
	if argument == "enable" {
		if len(payload) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Command 'enable' expects two arguments: <EventID> <channel>")
			return
		}
		payload[1] = CleanChannel(payload[1])
		err := h.EnableEvent(payload[0], payload[1], "")
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error enabling event: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Event enabled.")
		return
	}
	if argument == "disable" {
		if len(payload) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Command 'disable' expects two arguments: <EventID> <channel>")
			return
		}
		err := h.DisableEvent(payload[0], payload[1], "")
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error unwatching event: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Event disabled for: "+payload[1])
		return
	}
	if argument == "list" {
		formatted, err := h.ListEvents()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error listing events: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Events: "+formatted)
		return
	}
	if argument == "listenabled" {
		if len(payload) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Command 'listenabled' expects an argument: <ChannelID>")
			return
		}
		payload[0] = CleanChannel(payload[0])
		formatted, err := h.ListEnabled(payload[0])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error listing enabled events: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Enabled Events for <#"+payload[0]+">: "+formatted)
		return
	}
	if argument == "script" {
		if len(payload) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Command 'script' expects an argument: <eventID>")
			return
		}
		script, err := h.EventToScript(payload[0])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error retrieving script: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Script for "+payload[0]+": ```\n"+script+"\n```\n")
		return
	}
	if argument == "info" {
		if len(payload) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Command 'info' expects an argument")
			return
		}
	}
}

// EnableEvent function
func (h *EventHandler) EnableEvent(eventID, channelID string, keyvalueid string) (err error) {
	channelID = CleanChannel(channelID)
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		return err
	}

	err = h.AddEventToRoom(event.ID, channelID)
	if err != nil {
		return err
	}

	err = h.LoadEvent(event.ID, keyvalueid)
	if err != nil {
		return err
	}
	return nil
}

// EventToScript function
func (h *EventHandler) EventToScript(eventID string) (script string, err error) {
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		return "", err
	}

	script, err = h.parser.EventToJSON(event)
	if err != nil {
		return "", err
	}
	return script, nil
}

// DisableEvent function
func (h *EventHandler) DisableEvent(eventID string, channelID string, keyvalueid string) (err error) {
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		return err
	}
	err = h.RemoveEventFromRoom(eventID, channelID)
	if err != nil {
		return err
	}
	h.UnWatchEvent(channelID, event.ID, keyvalueid)
	return nil
}

// RemoveEvent function
func (h *EventHandler) RemoveEvent(eventID string, userID string, s *discordgo.Session, channelID string) (err error) {
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		return err
	}

	user, err := h.user.GetUser(userID, s, channelID)
	if err != nil {
		return err
	}

	if user.ID == event.CreatorID || user.CheckRole("admin") {
		// Unwatch this event in all rooms
		for _, roomID := range event.Rooms {
			h.UnWatchEvent(roomID, eventID, "")
		}
		err = h.eventsdb.RemoveEventByID(eventID)
		if err != nil {
			return err
		}
	} else {
		return errors.New("You do not have permission to remove this event only the creator or an admin are allowed to")
	}
	return nil
}

// ListEvents function
func (h *EventHandler) ListEvents() (formatted string, err error) {
	events, err := h.eventsdb.GetAllEvents()
	if err != nil {
		return "", err
	}

	formatted = "```"

	for _, event := range events {
		channels := ""
		for i, channel := range event.Rooms {
			if i == 0 {
				channels = channel
			} else {
				channels = channels + ", " + channel
			}
		}
		formatted = formatted + "\n" + event.ID + " - " + event.Name + ": " + event.Description
	}
	formatted = formatted + "\n```\n"
	return formatted, nil
}

// EventInfo function
func (h *EventHandler) EventInfo(eventID string) (formatted string, err error) {
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		return "", err
	}

	formatted = "```"
	channels := ""
	for i, channel := range event.Rooms {
		if i == 0 {
			channels = channel
		} else {
			channels = channels + ", " + channel
		}
	}
	formatted = formatted + "\nEventID: " + event.ID
	formatted = formatted + "\nName: " + event.Name
	formatted = formatted + "\nDescription: " + event.Description
	formatted = formatted + "\nChannelID: " + channels
	formatted = formatted + "\nType: " + event.Type
	formatted = formatted + "\nCreatorID: " + event.CreatorID
	formatted = formatted + "\n```\n"
	return formatted, nil
}

// RegisterEvent function
func (h *EventHandler) RegisterEvent(payload string, s *discordgo.Session, m *discordgo.MessageCreate) (eventID string, err error) {
	payload = strings.TrimPrefix(payload, "~events add ") // This all will need to be updated later, this is just
	payload = strings.TrimPrefix(payload, "\n")           // A lazy way of cleaning the command
	payload = strings.TrimPrefix(payload, "```")
	payload = strings.TrimPrefix(payload, "\n")
	payload = strings.TrimSuffix(payload, "\n")
	payload = strings.TrimSuffix(payload, "```")
	payload = strings.TrimSuffix(payload, "\n")
	payload = strings.Trim(payload, "```")

	createdEvent, err := h.parser.ParseFormattedEvent(payload, m.Author.ID)
	if err != nil {
		return "", err
	}

	id := strings.Split(GetUUIDv2(), "-")
	createdEvent.ID = id[0]

	// After our json is parsed, we need to validate the event to make sure it will run correctly
	err = h.parser.ValidateEvent(createdEvent)
	if err != nil {
		return "", err
	}

	err = h.eventsdb.SaveEventToDB(createdEvent)
	if err != nil {
		return "", err
	}
	return createdEvent.ID, nil
}

// LoadEventsAtBoot function
func (h *EventHandler) LoadEventsAtBoot() (err error) {
	events, err := h.eventsdb.GetAllEvents()
	if err != nil {
		return err
	}

	for _, event := range events {
		if event.LoadOnBoot {
			err = h.LoadEvent(event.ID, "")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// AddEventToRoom function
func (h *EventHandler) AddEventToRoom(eventID string, roomID string) (err error) {
	roomID = CleanChannel(roomID)

	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		return err
	}

	for _, room := range event.Rooms {
		if room == roomID {
			return nil // We don't want to add a record more than once
		}
	}
	event.Rooms = append(event.Rooms, roomID)

	err = h.eventsdb.SaveEventToDB(event)
	if err != nil {
		return err
	}
	return nil
}

// RemoveEventFromRoom function
func (h *EventHandler) RemoveEventFromRoom(eventID string, roomID string) (err error) {
	roomID = CleanChannel(roomID)

	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		return err
	}

	event.Rooms = RemoveStringFromSlice(event.Rooms, roomID)

	err = h.eventsdb.SaveEventToDB(event)
	if err != nil {
		return err
	}
	return nil
}

// LoadEvent function
func (h *EventHandler) LoadEvent(eventID string, keyvalueid string) (err error) {
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		return err
	}

	// Refer to the github wiki page on Events for information on types
	for _, room := range event.Rooms {
		err = h.AddEventToWatchList(event, room, keyvalueid)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddEventToWatchList function
func (h *EventHandler) AddEventToWatchList(event Event, roomID string, keyvalueid string) (err error) {
	if event.Type == "ReadMessage" {
		h.WatchEvent(h.UnfoldReadMessageEvent, keyvalueid, event.ID, roomID)
	} else if event.Type == "TimedMessage" {
		h.WatchEvent(h.UnfoldTimedMessageEvent, keyvalueid, event.ID, roomID)
	} else if event.Type == "ReadMessageChoice" {
		h.WatchEvent(h.UnfoldReadMessageChoiceEvent, keyvalueid, event.ID, roomID)
	} else if event.Type == "MessageChoiceTriggerEvent" {
		h.WatchEvent(h.UnfoldMessageChoiceTriggerEvent, keyvalueid, event.ID, roomID)
	}
	return nil
}

// WatchEvent function
func (h *EventHandler) WatchEvent(Handler func(string, string, *discordgo.Session, *discordgo.MessageCreate), KeyValueID string, EventID string, ChannelID string) {
	// Make sure we don't duplicate events in the watch list
	for e := h.WatchList.Front(); e != nil; e = e.Next() {
		r := reflect.ValueOf(e.Value)
		keyvalueID := reflect.Indirect(r).FieldByName("KeyValueID")
		eventID := reflect.Indirect(r).FieldByName("EventID")
		channel := reflect.Indirect(r).FieldByName("ChannelID")
		if channel.String() == ChannelID && eventID.String() == EventID && keyvalueID.String() == KeyValueID {
			return
		}
	}
	item := EventCallback{ChannelID: ChannelID, EventID: EventID, Handler: Handler, KeyValueID: KeyValueID}
	h.WatchList.PushBack(item)
}

// UnWatchEvent function
func (h *EventHandler) UnWatchEvent(ChannelID string, EventID string, KeyValueID string) {
	// Clear usermanager element by iterating
	for e := h.WatchList.Front(); e != nil; e = e.Next() {
		r := reflect.ValueOf(e.Value)
		keyvalueID := reflect.Indirect(r).FieldByName("KeyValueID")
		channel := reflect.Indirect(r).FieldByName("ChannelID")
		eventID := reflect.Indirect(r).FieldByName("EventID")

		if channel.String() == ChannelID && eventID.String() == EventID && keyvalueID.String() == KeyValueID {
			h.WatchList.Remove(e)
			h.RemoveEventFromRoom(EventID, ChannelID)
		}
	}
}

// ListEnabled function
func (h *EventHandler) ListEnabled(channelID string) (formatted string, err error) {
	formatted = "```\n"
	for e := h.WatchList.Front(); e != nil; e = e.Next() {
		r := reflect.ValueOf(e.Value)
		keyvalueID := reflect.Indirect(r).FieldByName("KeyValueID")
		eventID := reflect.Indirect(r).FieldByName("EventID")
		channel := reflect.Indirect(r).FieldByName("ChannelID")
		if channel.String() == channelID {
			formatted = formatted + "ID: " + eventID.String() + " KeyValueID: " + keyvalueID.String() + "\n"
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

			keyvalue := reflect.Indirect(r).FieldByName("KeyValueID")
			keyvalueid := keyvalue.String()

			// We now type the interface to the handler type
			//v := reflect.ValueOf(handler)
			rargs := make([]reflect.Value, 4)

			//var sizeofargs = len(rargs)
			rargs[0] = reflect.ValueOf(eventid)
			rargs[1] = reflect.ValueOf(keyvalueid)
			rargs[2] = reflect.ValueOf(s)
			rargs[3] = reflect.ValueOf(m)

			go handler.Call(rargs)
			//handlerid := reflect.Indirect(r).FieldByName("HandlerID").String()
			//c.UnWatchEvent(m.ChannelID, handlerid)
		}
	}
}

// CreateAttachedEvent function
func (h *EventHandler) CreateAttachedEvent(event Event, userID string) (attachedevent Event, err error) {
	// If we didn't find an event we need to save a new one in the db for this user
	event.UserAttached = event.ID + "-" + userID // We key the new event with the root Event ID and the User ID
	id := strings.Split(GetUUIDv2(), "-")
	event.ID = id[0] // Give this new event a new ID or it will overwrite the root record
	event.Name = event.Name + "-Attached"
	err = h.eventsdb.SaveEventToDB(event)
	if err != nil {
		return attachedevent, err
	}
	return event, nil
}

// UnfoldReadMessageEvent function
func (h *EventHandler) UnfoldReadMessageEvent(eventID string, keyvalueid string, s *discordgo.Session, m *discordgo.MessageCreate) {
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

	// We need to determine if the event is attachable or not first
	if event.Attachable {
		// If the event is attachable, then this is not the event we want to trigger, we want to retrieve the users attached event
		event, err = h.eventsdb.GetEventByAttached(event.ID, m.Author.ID)
		if err != nil {
			return // If we didn't find a record, one wasn't registered and we want to silently fail
		}
	} // Now we have the event attached to the user and can proceed with parsing it

	keyword := event.TypeFlags[0]
	messageContent := strings.Fields(strings.ToLower(m.Content))

	for _, messagefield := range messageContent {
		// We don't need to check for the userID here because that's what checking for event.Attachable did
		if messagefield == keyword {
			// First we send the data
			s.ChannelMessageSend(m.ChannelID, FormatEventMessage(event.Data[0], m.Author.ID, m.ChannelID))

			// We need to check if the cycles are indefinite or not
			h.CheckCycles(event, keyvalueid, s, m)
			return
		}
	}
	return
}

// UnfoldTimedMessageEvent function
func (h *EventHandler) UnfoldTimedMessageEvent(eventID string, keyvalueid string, s *discordgo.Session, m *discordgo.MessageCreate) {
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

	// We need to determine if the event is attachable or not first
	if event.Attachable {
		// If the event is attachable, then this is not the event we want to trigger, we want to retrieve the users attached event
		event, err = h.eventsdb.GetEventByAttached(event.ID, m.Author.ID)
		if err != nil {
			return // If we didn't find a record, one wasn't registered and we want to silently fail
		}
	} // Now we have the event attached to the user and can proceed with parsing it

	keyword := event.TypeFlags[0]
	timeout, _ := strconv.Atoi(event.TypeFlags[1]) // We don't bother checking for an error here because that was handled during the event registration.
	messageContent := strings.Fields(strings.ToLower(m.Content))

	for _, messagefield := range messageContent {
		if messagefield == keyword {
			// First we want to sleep for our timeout period
			time.Sleep(time.Duration(timeout) * time.Second)
			// Now we send the data
			s.ChannelMessageSend(m.ChannelID, FormatEventMessage(event.Data[0], m.Author.ID, m.ChannelID))

			// We need to check if the cycles are indefinite or not
			h.CheckCycles(event, keyvalueid, s, m)
			return
		}
	}
	return
}

// UnfoldReadMessageChoiceEvent function
func (h *EventHandler) UnfoldReadMessageChoiceEvent(eventID string, keyvalueid string, s *discordgo.Session, m *discordgo.MessageCreate) {
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

	// We need to determine if the event is attachable or not first
	if event.Attachable {
		// If the event is attachable, then this is not the event we want to trigger, we want to retrieve the users attached event
		event, err = h.eventsdb.GetEventByAttached(event.ID, m.Author.ID)
		if err != nil {
			return // If we didn't find a record, one wasn't registered and we want to silently fail
		}
	} // Now we have the event attached to the user and can proceed with parsing it

	messageContent := strings.Fields(strings.ToLower(m.Content))

	for i, field := range event.TypeFlags {
		for _, message := range messageContent {
			if field == message {
				// First we send the data that is keyed to the field
				s.ChannelMessageSend(m.ChannelID, FormatEventMessage(event.Data[i], m.Author.ID, m.ChannelID))

				// We need to check if the cycles are indefinite or not
				h.CheckCycles(event, keyvalueid, s, m)
				return
			}
		}
	}
}

// LaunchChildEvent function
func (h *EventHandler) LaunchChildEvent(parenteventID string, childeventID string, keyvalueid string, s *discordgo.Session, m *discordgo.MessageCreate) {
	// First we load the keyed eventID in the data array
	if childeventID != "nil" {
		triggeredevent, err := h.eventsdb.GetEventByID(childeventID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error triggering keyed event - Source: "+
				parenteventID+" Trigger: "+childeventID+" Error: "+err.Error())
			return
		}

		if triggeredevent.Watchable {
			err = h.AddEventToWatchList(triggeredevent, m.ChannelID, keyvalueid)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error watching keyed event - Source: "+
					parenteventID+" Trigger: "+childeventID+" Error: "+err.Error())
				return
			}
		} else {
			// If we aren't adding the event to the watchlist, we want to passthrough and trigger it immediately with a 2 second time delay
			time.Sleep(time.Duration(time.Second * 2))
			if triggeredevent.Type == "ReadMessage" {
				h.UnfoldReadMessageEvent(triggeredevent.ID, keyvalueid, s, m)
			} else if triggeredevent.Type == "TimedMessage" {
				h.UnfoldTimedMessageEvent(triggeredevent.ID, keyvalueid, s, m)
			} else if triggeredevent.Type == "ReadMessageChoice" {
				h.UnfoldReadMessageChoiceEvent(triggeredevent.ID, keyvalueid, s, m)
			} else if triggeredevent.Type == "MessageChoiceTriggerEvent" {
				h.UnfoldMessageChoiceTriggerEvent(triggeredevent.ID, keyvalueid, s, m)
			}
		}
	}
}

// CheckCycles function
func (h *EventHandler) CheckCycles(event Event, keyvalueid string, s *discordgo.Session, m *discordgo.MessageCreate) {
	// We need to check if the cycles are indefinite or not
	if event.Cycles > 0 {
		// We increment our run count and save the event to the db
		event.RunCount = event.RunCount + 1
		err := h.eventsdb.SaveEventToDB(event)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error saving event: "+event.ID+" Error: "+err.Error())
			return
		}
		// Then we check to see if we hit our cycle limit and if so then remove the event from the db
		if event.RunCount >= event.Cycles {
			// If this is an attached event we need to remove it from the DB after cleanup
			if event.UserAttached != "" {
				err = h.eventsdb.RemoveEventFromDB(event)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Error removing user attached event: "+event.ID+" Error: "+err.Error())
					return
				}
			} else {
				// Then we need to clear the runcount and unwatch it instead of deleting it
				event.RunCount = 0
				err = h.eventsdb.SaveEventToDB(event)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Error saving event: "+event.ID+" Error: "+err.Error())
					return
				}
			}
			h.UnWatchEvent(m.ChannelID, event.ID, keyvalueid)
			return
		}
	}
}

// IsValidEventMessage function
func (h *EventHandler) IsValidEventMessage(s *discordgo.Session, m *discordgo.MessageCreate) (valid bool) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return false
	}
	// Ignore bots
	if m.Author.Bot {
		return false
	}
	return true
}

// UnfoldMessageChoiceTriggerEvent function
func (h *EventHandler) UnfoldMessageChoiceTriggerEvent(eventID string, keyvalueid string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if !h.IsValidEventMessage(s, m) {
		return
	}

	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading event "+eventID+": Error: "+err.Error())
		return
	}

	// We need to determine if the event is attachable or not first
	if event.Attachable {
		// If the event is attachable, then this is not the event we want to trigger, we want to retrieve the users attached event
		event, err = h.eventsdb.GetEventByAttached(event.ID, m.Author.ID)
		if err != nil {
			return // If we didn't find a record, one wasn't registered and we want to silently fail
		}
	} // Now we have the event attached to the user and can proceed with parsing it

	messageContent := strings.Fields(strings.ToLower(m.Content))

	for i, field := range event.TypeFlags {
		for _, message := range messageContent {
			if field == message {
				// First we load the keyed eventID in the data array
				if event.Data[i] != "nil" {
					//fmt.Println("Launching child event parent - " + event.ID + " - childeventID: " + event.Data[i] + " keyvalue: " + keyvalueid )
					go h.LaunchChildEvent(event.ID, event.Data[i], keyvalueid, s, m)
				}
				// We need to check if the cycles are indefinite or not
				h.CheckCycles(event, keyvalueid, s, m)
				return
			}
		}
	}
}
