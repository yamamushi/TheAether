package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"fmt"
	"reflect"
	"container/list"
	"errors"
)

// This handles the discord input interface for registering events and managing them

// EventHandler struct
type EventHandler struct {

	conf       *Config
	registry   *CommandRegistry
	callback   *CallbackHandler
	db         *DBHandler
	user	   *UserHandler

	WatchList list.List
	dg        *discordgo.Session
	logger    *Logger

	eventsdb	*EventsDB
	parser 		*EventParser

}

// EventCallback struct
type EventCallback struct {
	ChannelID string
	EventID   string
	Handler   func(string, *discordgo.Session, *discordgo.MessageCreate)
}

// Init function
func (h *EventHandler) Init() (err error) {
	fmt.Println("Registering Event Handler Command")
	h.eventsdb = new(EventsDB)
	h.eventsdb.db = h.db
	h.parser = new(EventParser)
	h.RegisterCommand()

	fmt.Println("Loading Registered Events from Database")
	err = h.LoadEvents()
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
			s.ChannelMessageSend(m.ChannelID, "Error registering new event: " + err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Event registered with ID: " + eventID)
		return
	}
	if argument == "remove" {
		if len(payload) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Command 'remove' expects an argument")
			return
		}
		err := h.RemoveEvent(payload[0], m.Author.ID, s, m.ChannelID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error removing event: " + err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Event record removed!")
		return
	}
	if argument == "enable" {
		if len(payload) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Command 'enable' expects an argument: <EventID>")
			return
		}
		payload[0] = CleanChannel(payload[0])
		err := h.EnableEvent(payload[0], m.Author.ID, m.ChannelID,  s)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error enabling event: " + err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Event enabled.")
		return
	}
	if argument == "disable" {
		if len(payload) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Command 'disable' expects an argument	: <EventID>")
			return
		}
		err := h.DisableEvent(payload[0],m.Author.ID,s)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error unwatched event: " + err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Event removed from watchlist!")
		return
	}
	if argument == "list" {
		formatted,err := h.ListEvents()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error listing events: " + err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Events: " +formatted)
		return
	}
	if argument == "listenabled" {
		if len(payload) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Command 'listenabled' expects an argument: <ChannelID>")
			return
		}
		payload[0] = CleanChannel(payload[0])
		formatted,err := h.ListEnabled(payload[0])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error listing enabled events: " + err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Enabled Events for <#"+payload[0]+">: " +formatted)
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
func (h *EventHandler) EnableEvent(eventID, userID string, channelID string, s *discordgo.Session) (err error){
	user, err := h.user.GetUser(userID, s, channelID)
	if err != nil {
		return err
	}

	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		return err
	}

	if user.ID == event.CreatorID || user.CheckRole("builder") {
		h.LoadEvent(eventID)
	} else {
		return errors.New("You do not have permission to enable this event only the creator or a builder are allowed to")
	}
	return nil

}

// DisableEvent function
func (h *EventHandler) DisableEvent(eventID string, userID string, s *discordgo.Session) (err error){
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		return err
	}

	user, err := h.user.GetUser(userID, s, event.ChannelID)
	if err != nil {
		return err
	}

	if user.ID == event.CreatorID || user.CheckRole("builder") {
		h.UnWatchEvent(event.ChannelID, event.ID)
	} else {
		return errors.New("You do not have permission to disable this event only the creator or a builder are allowed to")
	}
	return nil
}

// RemoveEvent function
func (h *EventHandler) RemoveEvent(eventID string, userID string, s *discordgo.Session, channelID string) (err error){
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		return err
	}

	user, err := h.user.GetUser(userID, s, channelID)
	if err != nil {
		return err
	}

	if user.ID == event.CreatorID || user.CheckRole("admin") {
		err = h.eventsdb.RemoveEventByID(eventID)
		if err != nil {
			return err
		}
		h.UnWatchEvent(channelID, eventID)
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

	formatted = "```\n"

	for _, event := range events {
		formatted = formatted + "EventID: " +event.ID + " ChannelID:" + event.ChannelID + " Type:" + event.Type + " CreatorID:" + event.CreatorID + "\n"
	}
	formatted = formatted + "\n```\n"
	return formatted, nil
}


// RegisterEvent function
func (h *EventHandler) RegisterEvent(payload string, s *discordgo.Session, m *discordgo.MessageCreate) (eventID string, err error){
	payload = strings.TrimPrefix(payload, "~events add ")
	payload = strings.TrimPrefix(payload, "\n")
	payload = strings.TrimPrefix(payload, "```")
	payload = strings.TrimPrefix(payload, "\n")
	payload = strings.TrimSuffix(payload, "\n")
	payload = strings.TrimSuffix(payload, "```")
	payload = strings.TrimSuffix(payload, "\n")
	payload = strings.Trim(payload, "```")

	createdEvent, err := h.parser.ParseFormattedEvent(payload, m.ChannelID, m.Author.ID)
	if err != nil {
		return "", err
	}

	id := strings.Split(GetUUIDv2(), "-")
	createdEvent.ID = id[0]

	err = h.eventsdb.SaveEventToDB(createdEvent)
	if err != nil {
		return "", err
	}
	return createdEvent.ID, nil
}


// LoadEvents function
func (h *EventHandler) LoadEvents() (err error){
	events, err := h.eventsdb.GetAllEvents()
	if err != nil {
		return err
	}

	for _, event := range events {
		err = h.LoadEvent(event.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// LoadEvent function
func (h *EventHandler) LoadEvent(eventID string) (err error){
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		return err
	}

	if event.Type == "ReadMessage"{
		//TypeFields
		// 0 - word to trigger on
		h.WatchEvent(h.UnfoldReadMessageEvent, event.ID, event.ChannelID)
	} else if event.Type == "TimedMessage"{
		//TypeFields
		// 0 - Formatted Duration String
		// Data - Formatted message to send
		//h.Watchevent(h.UnfoldTimedMessageEvent, event.ID, event.ChannelID)
	}
	return nil
}




// WatchEvent function
func (h *EventHandler) WatchEvent(Handler func(string, *discordgo.Session, *discordgo.MessageCreate), EventID string, ChannelID string) {
	for e := h.WatchList.Front(); e != nil; e = e.Next() {
		r := reflect.ValueOf(e.Value)
		eventID := reflect.Indirect(r).FieldByName("EventID")
		channel := reflect.Indirect(r).FieldByName("ChannelID")
		if channel.String() == ChannelID && eventID.String() == EventID {
			return
		}
	}
	item := EventCallback{ChannelID: ChannelID, EventID: EventID, Handler: Handler}
	h.WatchList.PushBack(item)
}

// UnWatchEvent function
func (h *EventHandler) UnWatchEvent(ChannelID string, EventID string) {
	// Clear usermanager element by iterating
	for e := h.WatchList.Front(); e != nil; e = e.Next() {
		r := reflect.ValueOf(e.Value)
		channel := reflect.Indirect(r).FieldByName("ChannelID")
		eventID := reflect.Indirect(r).FieldByName("EventID")

		if channel.String() == ChannelID && eventID.String() == EventID {
			h.WatchList.Remove(e)
		}
	}
}

// ListEnabled function
func (h *EventHandler) ListEnabled(channelID string) (formatted string, err error){
	formatted = "```\n"
	for e := h.WatchList.Front(); e != nil; e = e.Next() {
		r := reflect.ValueOf(e.Value)
		eventID := reflect.Indirect(r).FieldByName("EventID")
		channel := reflect.Indirect(r).FieldByName("ChannelID")
		if channel.String() == channelID{
			formatted = formatted +"ID: "+ eventID.String() + "\n"
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

			// We now type the interface to the handler type
			//v := reflect.ValueOf(handler)
			rargs := make([]reflect.Value, 3)

			//var sizeofargs = len(rargs)
			rargs[0] = reflect.ValueOf(eventid)
			rargs[1] = reflect.ValueOf(s)
			rargs[2] = reflect.ValueOf(m)

			go handler.Call(rargs)
			//handlerid := reflect.Indirect(r).FieldByName("HandlerID").String()
			//c.UnWatchEvent(m.ChannelID, handlerid)
		}
	}
}

// UnfoldReadMessageEvent function
func (h *EventHandler) UnfoldReadMessageEvent(eventID string, s *discordgo.Session, m *discordgo.MessageCreate) {
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
		s.ChannelMessageSend(m.ChannelID, "Error loading event: " + eventID + " Error: " + err.Error())
		return
	}

	eventstring := ""
	if len(event.TypeFlags) > 0 {
		eventstring = event.TypeFlags[0] // The first value in the typeflag for a readmessage event is the string to parse for
	}

	messageContent := strings.Fields(strings.ToLower(m.Content))

	for _, messagefield := range messageContent {
		if messagefield == eventstring {
			s.ChannelMessageSend(event.ChannelID, event.Data)
			return
		}
	}
	return
}