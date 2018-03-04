package main

import (
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"strings"
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

	eventsdb      *EventsDB
	parser        *EventParser
	eventmessages *EventMessagesDB
}

// Init function
func (h *EventHandler) Init() (err error) {
	fmt.Println("Registering Event Handler Command")
	h.eventsdb = new(EventsDB)
	h.eventsdb.db = h.db
	h.parser = new(EventParser)
	h.parser.eventsdb = h.eventsdb
	h.RegisterCommand()

	//fmt.Println("Loading Registered Events from Database")
	//err = h.LoadEventsAtBoot()
	//if err != nil {
	//	return err
	//}
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
		err := h.EnableEvent(payload[0], payload[1], "", s, m)
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
		h.UnWatchEvent(payload[1], payload[0], "")
		err := h.DisableEvent(payload[0], payload[1])
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
		script, err := h.EventToJSONString(payload[0])
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
	if argument == "update" {
		if len(payload) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Command 'update' expects two arguments: <eventID> <payload>")
			return
		}
		// We pass in the full message here because we intend to unpack it later
		err := h.UpdateEvent(payload[0], m.Content, s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error registering new event: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Event updated")
		return
	}
	if argument == "loadfromdisk" {
		force := false
		if len(payload) > 0 {
			if payload[0] == "force" {
				force = true
			}
		}
		err := h.LoadEventsFromDisk(force)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error loading events from disk: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Events loaded from disk")
		return
	}
	if argument == "savetodisk" {
		err := h.SaveEvents()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error saving events: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Events saved to disk")
		return
	}
	if argument == "save" {
		if len(payload) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Command 'save' expects an argument: <eventID>")
			return
		}
		err := h.SaveEventToDisk(payload[0])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error saving event: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Event saved to disk")
		return
	}
}

// EnableEvent function
func (h *EventHandler) EnableEvent(eventID, channelID string, keyvalueid string, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {
	channelID = CleanChannel(channelID)
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		return err
	}

	err = h.AddEventToRoom(event.ID, channelID)
	if err != nil {
		return err
	}

	err = h.LoadEvent(event.ID, keyvalueid, s, m)
	if err != nil {
		return err
	}
	return nil
}

// SaveEvents function
func (h *EventHandler) SaveEvents() (err error) {
	events, err := h.eventsdb.GetAllEvents()
	if err != nil {
		return err
	}
	CreateDirIfNotExist("events")

	for _, event := range events {
		if !event.IsScriptEvent {
			err = h.SaveEventToDisk(event.ID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// LoadEventsFromDisk function
func (h *EventHandler) LoadEventsFromDisk(force bool) (err error) {
	files, err := ioutil.ReadDir("events/")
	if err != nil {
		return errors.New("Error reading directory: " + err.Error())
	}

	for _, file := range files {
		if strings.Contains(file.Name(), ".event") {
			data, err := ioutil.ReadFile("events/" + file.Name())
			if err != nil {
				return errors.New("Error reading file: " + file.Name() + " - " + err.Error())
			}

			event, err := h.UnmarshalEvent(data)
			if err != nil {
				return errors.New("Error unpacking file: " + file.Name() + " - " + err.Error())
			}

			if force {
				// Remove then save to disk
				h.eventsdb.RemoveEventByID(event.ID)
				err = h.eventsdb.SaveEventToDB(event)
				if err != nil {
					return err
				}
			} else {
				// Only save if doesn't exist already
				_, err = h.eventsdb.GetEventByID(event.ID)
				if err != nil {
					err = h.eventsdb.SaveEventToDB(event)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// SaveEventToDisk function
func (h *EventHandler) SaveEventToDisk(eventID string) (err error) {
	formattedjson, err := h.EventToJSONString(eventID)
	if err != nil {
		return err
	}
	_ = ioutil.WriteFile("events/"+eventID+".event", []byte(formattedjson), 0644)
	return nil
}

// EventToJSONString function
func (h *EventHandler) EventToJSONString(eventID string) (script string, err error) {
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
func (h *EventHandler) DisableEvent(eventID string, channelID string) (err error) {
	channelID = CleanChannel(channelID)
	_, err = h.eventsdb.GetEventByID(eventID)
	if err != nil {
		return err
	}
	err = h.RemoveEventFromRoom(eventID, channelID)
	if err != nil {
		return err
	}
	return nil
}

// RemoveEvent function
func (h *EventHandler) RemoveEvent(eventID string, userID string, s *discordgo.Session, channelID string) (err error) {
	channelID = CleanChannel(channelID)

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
		// We don't want to see events in scripts
		if !event.IsScriptEvent {
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

// UpdateEvent function
func (h *EventHandler) UpdateEvent(eventID string, payload string, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {

	payload = strings.TrimPrefix(payload, "~events update "+eventID+" ") // This all will need to be updated later, this is just
	payload = strings.TrimPrefix(payload, "\n")                          // A lazy way of cleaning the command
	payload = strings.TrimPrefix(payload, "```")
	payload = strings.TrimPrefix(payload, "\n")
	payload = strings.TrimSuffix(payload, "\n")
	payload = strings.TrimSuffix(payload, "```")
	payload = strings.TrimSuffix(payload, "\n")
	payload = strings.Trim(payload, "```")

	createdEvent, err := h.parser.VerifyUpdateEvent(payload, m.Author.ID)
	if err != nil {
		return err
	}

	// After our json is parsed, we need to validate the event to make sure it will run correctly
	err = h.parser.ValidateEvent(createdEvent)
	if err != nil {
		return err
	}

	err = h.eventsdb.SaveEventToDB(createdEvent)
	if err != nil {
		return err
	}
	return nil
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
			/*err = h.LoadEvent(event.ID, "")
			if err != nil {
				return err
			}
			*/
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
func (h *EventHandler) LoadEvent(eventID string, keyvalueid string, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {
	event, err := h.eventsdb.GetEventByID(eventID)
	if err != nil {
		return err
	}

	// Refer to the github wiki page on Events for information on types
	for _, room := range event.Rooms {
		h.LaunchChildEvent("", event.ID, keyvalueid, room, s, m)
	}
	return nil
}

// CheckCycles function
/*
func (h *EventHandler) CheckCycles(event Event, eventmessagesid string, s *discordgo.Session, m *discordgo.MessageCreate) {
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
			// Now we disable the event if it is past the run cycle count
			h.UnWatchEvent(m.ChannelID, event.ID, eventmessagesid)
			err = h.DisableEvent(event.ID, m.ChannelID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error disabling event: "+event.ID+" Error: "+err.Error())
				return
			}
			// Reset the runcount and save it to disk
			event.RunCount = 0
			err = h.eventsdb.SaveEventToDB(event)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error saving event after cycles: "+event.ID+" Error: "+err.Error())
				return
			}
			return
		}
	}
}
*/

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

// UnmarshalEvent function
func (h *EventHandler) UnmarshalEvent(data []byte) (event Event, err error) {
	if err := json.Unmarshal(data, &event); err != nil {
		return event, err
	}
	return event, nil
}
