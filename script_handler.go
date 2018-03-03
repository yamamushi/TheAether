package main

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strconv"
	"strings"
)

// ScriptHandler struct
type ScriptHandler struct {
	scriptsdb       *ScriptsDB
	db              *DBHandler
	registry        *CommandRegistry
	conf            *Config
	eventhandler    *EventHandler
	eventmessagesdb *EventMessagesDB
}

// Init function
func (h *ScriptHandler) Init() (err error) {
	fmt.Println("Registering Script Handler Command")
	h.scriptsdb = new(ScriptsDB)
	h.scriptsdb.db = h.db
	h.RegisterCommand()
	return nil
}

// RegisterCommand command function
func (h *ScriptHandler) RegisterCommand() {
	h.registry.Register("script", "Manage scripts", "add|remove|edit|search|info")
	h.registry.AddGroup("script", "builder")
}

// Read function
func (h *ScriptHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {
	cp := h.conf.MainConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		return
	}

	if strings.HasPrefix(m.Content, cp+"script") {
		if h.registry.CheckPermission("script", user, s, m) {

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
func (h *ScriptHandler) ParseCommand(input []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	command, arguments := GetArgumentAndFlags(input) // Okay so the name of this should probably be changed now

	if command == "add" {
		if len(arguments) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Command 'add' expects two arguments: <name> <description>")
			return
		}
		description := ""
		for i, field := range arguments {
			if i > 0 {
				if i == 1 {
					description = field
				} else {
					description = description + " " + field
				}
			}
		}

		scriptID, err := h.AddScript(arguments[0], m.Author.ID, description)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error registering new script: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Script `"+arguments[0]+"` registered with ID: "+scriptID)
		return
	}
	if command == "init" {
		if len(arguments) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Command 'init' expects two arguments: <scriptName> <rootEventID>")
			return
		}
		err := h.CreateExecutable(arguments[0], arguments[1])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error setting root eventID: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Script `"+arguments[0]+"` initialized.")
		return
	}
	if command == "test" {
		if len(arguments) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Command 'test' expects an arguments: <scriptName>")
			return
		}
		_, err := h.ExecuteScript(arguments[0], s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error executing script: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Script executed successfully: "+arguments[0])
		return
	}
	if command == "remove" {
		if len(arguments) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Command 'remove' expects an arguments: <scriptName>")
			return
		}
		err := h.RemoveScript(arguments[0])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error removing script: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Script Removed")
		return
	}
	if command == "list" {
		scriptlist, err := h.GetFormattedScriptList()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error retrieving script list: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Script list: \n"+scriptlist)
		return
	}
	/*
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
	   if argument == "search" {
	       formatted, err := h.ListEvents()
	       if err != nil {
	           s.ChannelMessageSend(m.ChannelID, "Error listing events: "+err.Error())
	           return
	       }
	       s.ChannelMessageSend(m.ChannelID, "Events: "+formatted)
	       return
	   }
	   if argument == "info" {
	       if len(payload) < 1 {
	           s.ChannelMessageSend(m.ChannelID, "Command 'info' expects an argument")
	           return
	       }
	   }
	*/
}

// GetFormattedScriptList function
func (h *ScriptHandler) GetFormattedScriptList() (formattedlist string, err error) {
	scriptslist, err := h.scriptsdb.GetAllScripts()
	if err != nil {
		return formattedlist, err
	}

	formattedlist = "```\n "
	for _, script := range scriptslist {
		formattedlist = formattedlist + "Name: " + script.Name + "\n "
		formattedlist = formattedlist + "ID: " + script.ID + "\n "
		formattedlist = formattedlist + "Desc: " + script.Description + "\n "
		formattedlist = formattedlist + "Creator: " + script.CreatorID + "\n "
		formattedlist = formattedlist + "EventMsgID: " + script.EventMessagesID + "\n "
		formattedlist = formattedlist + "Executable: " + strconv.FormatBool(script.Executable) + "\n "
		formattedlist = formattedlist + "------------------------\n "
	}
	formattedlist = formattedlist + "```\n "
	return formattedlist, nil
}

// AddScript function
func (h *ScriptHandler) AddScript(scriptName string, userID string, description string) (scriptID string, err error) {
	_, err = h.scriptsdb.GetScriptByName(scriptName)
	if err == nil {
		return scriptID, errors.New("Script with name " + scriptName + " already exists")
	}

	// Generate and assign an ID to this event
	scriptID = strings.Split(GetUUIDv2(), "-")[0]
	eventmessagesID := strings.Split(GetUUIDv2(), "-")[0]
	newscript := Script{ID: scriptID, Name: scriptName, CreatorID: userID, Description: description, EventMessagesID: eventmessagesID}

	err = h.scriptsdb.SaveScriptToDB(newscript)
	if err != nil {
		return scriptID, err
	}

	eventmessage := EventMessageContainer{ID: eventmessagesID, ScriptID: scriptID}
	err = h.eventmessagesdb.SaveEventMessageToDB(eventmessage)
	if err != nil {
		return scriptID, err
	}

	_, err = h.eventmessagesdb.GetEventMessageByID(eventmessagesID)
	if err != nil {
		return scriptID, err
	}
	return scriptID, nil
}

// CreateExecutable function
// This will overwrite any events in the script
func (h *ScriptHandler) CreateExecutable(scriptName string, startingeventID string) (err error) {

	//fmt.Println("Clone events")
	err = h.CloneEvents(scriptName, startingeventID)
	if err != nil {
		return err
	}

	//fmt.Println("Repair Events")
	err = h.RepairEvents(scriptName)
	if err != nil {
		return err
	}

	script, err := h.scriptsdb.GetScriptByName(scriptName)
	if err != nil {
		return err
	}

	script.Executable = true
	err = h.scriptsdb.SaveScriptToDB(script)
	if err != nil {
		return err
	}
	return nil
}

// CloneEvents function
func (h *ScriptHandler) CloneEvents(scriptName string, rooteventID string) (err error) {
	script, err := h.scriptsdb.GetScriptByName(scriptName)
	if err != nil {
		return err
	}

	if script.Executable {
		return errors.New("Script is already executable and cannot be modified")
	}

	rootevent, err := h.eventhandler.eventsdb.GetEventByID(rooteventID)
	if err != nil {
		return err
	}

	// First we get the events from the root event
	eventlist, err := h.GetDataEvents(rootevent.ID)
	if err != nil {
		return err
	}
	if len(eventlist) > 0 {
		// While we have events in the list
		for len(eventlist) > 0 {
			// Iterate through each event in the list
			for _, eventinlist := range eventlist {
				h.AddEventToScriptList(eventinlist, script.Name)

				// Find each event in the data fields of the event in the list we are parsing
				foundevents, err := h.GetDataEvents(eventinlist)
				if err != nil {
					return err
				}
				// If we found any events, we add them to the list if not in the list
				if len(foundevents) > 0 {
					for _, found := range foundevents {
						eventlist = AppendIfMissingString(eventlist, found)
					}
				}
				// Now we remove the event we just searched from the list
				eventlist = RemoveStringFromSlice(eventlist, eventinlist)
			}
		}
	} else {
		script.EventIDs = append(script.EventIDs, rootevent.ID)
	}
	// Now that we have an event list, we want to clone it for our script

	var clonedEventList []string
	for _, eventID := range script.EventIDs {
		event, err := h.eventhandler.eventsdb.GetEventByID(eventID)
		if err != nil {
			return err
		}

		clonedEvent := event

		newEventID := strings.Split(GetUUIDv2(), "-")[0]
		clonedEvent.ID = newEventID
		clonedEvent.EventMessagesID = script.EventMessagesID
		clonedEvent.IsScriptEvent = true
		clonedEvent.OriginalID = event.ID

		err = h.eventhandler.eventsdb.SaveEventToDB(clonedEvent)
		if err != nil {
			return err
		}
		clonedEventList = append(clonedEventList, clonedEvent.ID)
	}

	script.EventIDs = clonedEventList
	err = h.scriptsdb.SaveScriptToDB(script)
	if err != nil {
		return err
	}
	return nil
}

// RepairEvents function
func (h *ScriptHandler) RepairEvents(scriptName string) (err error) {
	script, err := h.scriptsdb.GetScriptByName(scriptName)
	if err != nil {
		return err
	}

	// Now we want to repair the eventlist in our records
	for _, finalscripteventid := range script.EventIDs {
		eventslist, err := h.GetDataEvents(finalscripteventid)
		if err != nil {
			return err
		}
		if len(eventslist) > 0 {
			// We first grab the event we're on
			event, err := h.eventhandler.eventsdb.GetEventByID(finalscripteventid)
			if err != nil {
				return err
			}
			// The repair happens here
			// Now we go through the nested events
			for i, nestedeventid := range event.Data {
				// We need to find the corresponding event from the parent script list
				for _, scripteventid := range script.EventIDs {
					scriptevent, err := h.eventhandler.eventsdb.GetEventByID(scripteventid)
					if err != nil {
						return err
					}

					// If the nested eventID matches the event's original id, then we replace the
					// Nested ID with the new ID and save it
					if scriptevent.OriginalID == nestedeventid {
						event.Data[i] = scriptevent.ID
						err = h.eventhandler.eventsdb.SaveEventToDB(event)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

// RemoveScript function
func (h *ScriptHandler) RemoveScript(scriptName string) (err error) {
	script, err := h.scriptsdb.GetScriptByName(scriptName)
	if err != nil {
		return err
	}

	// While we are removing the script we don't want anyone to execute it
	script.Executable = false
	err = h.scriptsdb.SaveScriptToDB(script)
	if err != nil {
		return err
	}

	// We need disable and remove each cloned event
	for _, eventid := range script.EventIDs {
		event, err := h.eventhandler.eventsdb.GetEventByID(eventid)
		if err != nil {
			fmt.Println("Error retrieiving event by id")
			return err
		}

		for _, channelid := range event.Rooms {
			err = h.eventhandler.DisableEvent(eventid, channelid, event.EventMessagesID)
			if err != nil {
				fmt.Println("Disable event failure")
				return err
			}
		}
		err = h.eventhandler.eventsdb.RemoveEventFromDB(event)
		if err != nil {
			fmt.Println("Remove event failure")
			return err
		}
	}

	// We need to remove the event message, however one may not exist
	_ = h.eventmessagesdb.RemoveEventMessageByID(script.EventMessagesID)

	err = h.scriptsdb.RemoveScriptByID(script.ID)
	if err != nil {
		return err
	}
	return nil
}

// AddEventToScriptList function
func (h *ScriptHandler) AddEventToScriptList(eventID string, scriptName string) (err error) {
	script, err := h.scriptsdb.GetScriptByName(scriptName)
	if err != nil {
		return err
	}

	for _, eventinscript := range script.EventIDs {
		if eventinscript == eventID {
			return nil
		}
	}
	script.EventIDs = append(script.EventIDs, eventID)
	err = h.scriptsdb.SaveScriptToDB(script)
	if err != nil {
		return err
	}
	return nil
}

// GetDataEvents function
func (h *ScriptHandler) GetDataEvents(eventID string) (eventids []string, err error) {
	rootEvent, err := h.eventhandler.eventsdb.GetEventByID(eventID)
	if err != nil {
		return eventids, err
	}
	// Now begin our traversal
	if h.eventhandler.parser.HasEventsInData(rootEvent) {
		for _, dataeventid := range rootEvent.Data {
			if dataeventid != "nil" {
				datafieldevent, err := h.eventhandler.eventsdb.GetEventByID(dataeventid)
				if err != nil {
					return eventids, err
				}
				eventids = append(eventids, datafieldevent.ID)
			}
		}
	}
	return eventids, nil
}

// ExecuteScript function
// This is where the fun happens
func (h *ScriptHandler) ExecuteScript(scriptName string, s *discordgo.Session, m *discordgo.MessageCreate) (status bool, err error) {
	// First verify the script exists and that the execution paths are valid
	script, err := h.scriptsdb.GetScriptByName(scriptName)
	if err != nil {
		return false, err
	}

	if !script.Executable {
		return false, errors.New("Script is not executable ")
	}

	if len(script.EventIDs) <= 0 {
		return false, errors.New("script has not been setup yet")
	}

	rootEvent, err := h.eventhandler.eventsdb.GetEventByID(script.EventIDs[0])
	if err != nil {
		return false, err
	}

	// We want to clear the container
	err = h.eventmessagesdb.ClearEventMessage(script.EventMessagesID, script.ID)
	if err != nil {
		return false, err
	}

	//fmt.Println("parsing event: " + rootEvent.ID)
	if rootEvent.Watchable {
		//fmt.Println("Adding to watchlist: " + rootEvent.ID)
		err = h.eventhandler.AddEventToWatchList(rootEvent, m.ChannelID, script.EventMessagesID)
		if err != nil {
			return false, err
		}
	} else {
		//fmt.Println("Launching root event: " + rootEvent.ID)
		// If we aren't watching the event, we'd like to get the key value response from it
		h.eventhandler.LaunchChildEvent("RootEvent", rootEvent.ID, script.EventMessagesID, s, m)
	}

	return false, nil
}
