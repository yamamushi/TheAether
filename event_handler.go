package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"fmt"
	"reflect"
	"container/list"
	"time"
	"errors"
)

// This handles the discord input interface for registering events and managing them

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

}

// WatchEvent struct
type WatchEvent struct {
	ChannelID string
	EventID   string
	Handler   func(string, *discordgo.Session, *discordgo.MessageCreate)
}


func (h *EventHandler) Init() (err error) {
	fmt.Println("Registering Event Handler Command")
	h.eventsdb = new(EventsDB)
	h.eventsdb.db = h.db
	h.RegisterCommand()

	fmt.Println("Loading Registered Events from Database")
	err = h.LoadEvents()
	if err != nil {
		return err
	}
	return nil
}

func (h *EventHandler) RegisterCommand() {
	h.registry.Register("events", "Manage events", "add|remove|list|info")
	h.registry.AddGroup("events", "builder")

}

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
	if argument == "unwatch" {
		/*if len(payload) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Command 'unwatch' expects an argument")
			return
		}
		err := h.UnWatchEvent(m.ChannelID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error unwatched event: " + err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Event removed from watchlist!")
		return*/
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
	if argument == "info" {
		if len(payload) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Command 'info' expects an argument")
			return
		}

	}
}


func (h *EventHandler) RemoveEvent(eventID string, userID string, s *discordgo.Session, channelID string) (err error){
	h.UnWatchEvent(channelID, eventID)

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
	} else {
		return errors.New("You do not have permission to remove this event, only the creator or an admin are allowed to.")
	}
	return nil
}


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



func (h *EventHandler) RegisterEvent(payload string, s *discordgo.Session, m *discordgo.MessageCreate) (eventID string, err error){
	waittime := time.Duration(time.Second*5)
	var typeflags []string

	// For testing
	typeflags = append(typeflags, "hello")
	id := strings.Split(GetUUIDv2(), "-")
	createdEvent := Event{ID: id[0], CreatorID: m.Author.ID, Type: "ReadMessage", TypeFlags: typeflags,
								TimeDelay: waittime.String(), Data: "Hello", ChannelID: m.ChannelID}

	err = h.eventsdb.SaveEventToDB(createdEvent)
	if err != nil {
		return "", err
	}

	if createdEvent.Type == "ReadMessage"{
		h.WatchEvent(h.UnfoldInputEvent, createdEvent.ID, m.ChannelID)
	}
	return createdEvent.ID, nil
}



func (h *EventHandler) LoadEvents() (err error){
	events, err := h.eventsdb.GetAllEvents()
	if err != nil {
		return err
	}

	for _, event := range events {
		if event.Type == "ReadMessage"{
			h.WatchEvent(h.UnfoldInputEvent, event.ID, event.ChannelID)
		}
	}
	return nil
}





// Watch function
func (c *EventHandler) WatchEvent(Handler func(string, *discordgo.Session, *discordgo.MessageCreate),
	EventID string, ChannelID string) {

	item := WatchEvent{ChannelID: ChannelID, EventID: EventID, Handler: Handler}
	c.WatchList.PushBack(item)
}

// UnWatch function
func (c *EventHandler) UnWatchEvent(ChannelID string, EventID string) {
	// Clear usermanager element by iterating
	var next *list.Element
	for e := c.WatchList.Front(); e != nil; e = next {
		next = e.Next()

		r := reflect.ValueOf(e.Value)
		channel := reflect.Indirect(r).FieldByName("ChannelID")
		eventID := reflect.Indirect(r).FieldByName("EventID")

		if channel.String() == ChannelID && eventID.String() == EventID {
			c.WatchList.Remove(e)
		}
	}
}



// Read function
func (c *EventHandler) ReadEvents(s *discordgo.Session, m *discordgo.MessageCreate) {
	for e := c.WatchList.Front(); e != nil; e = e.Next() {
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



func (h *EventHandler) UnfoldInputEvent(eventID string, s *discordgo.Session, m *discordgo.MessageCreate) {
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
			s.ChannelMessageSend(event.ChannelID, event.Data + " " + m.Author.Mention() + "!")
			return
		}
	}
	return
}