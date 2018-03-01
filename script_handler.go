package main

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
)

// ScriptHandler struct
type ScriptHandler struct {
	scriptsdb    *ScriptsDB
	db           *DBHandler
	registry     *CommandRegistry
	conf         *Config
	eventhandler *EventHandler
	keyvaluesdb  *KeyValuesDB
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
	if command == "setroot" {
		if len(arguments) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Command 'add' expects two arguments: <scriptID> <rootEventID>")
			return
		}
		err := h.SetRootEvent(arguments[0], arguments[1])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error setting root eventID: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Script `"+arguments[0]+"` root eventID: "+arguments[1])
		return
	}
	if command == "test" {
		if len(arguments) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Command 'test' expects an arguments: <scriptID>")
			return
		}
		err := h.ExecuteScript(arguments[0], s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error executing script: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Script executed successfully: "+arguments[0])
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

// AddScript function
func (h *ScriptHandler) AddScript(scriptName string, userID string, description string) (scriptID string, err error) {
	_, err = h.scriptsdb.GetScriptByName(scriptName)
	if err == nil {
		return scriptID, errors.New("Script with name " + scriptName + " already exists")
	}

	// Generate and assign an ID to this event
	scriptID = strings.Split(GetUUIDv2(), "-")[0]
	newscript := Script{ID: scriptID, Name: scriptName, CreatorID: userID, Description: description}

	err = h.scriptsdb.SaveScriptToDB(newscript)
	if err != nil {
		return scriptID, err
	}
	return scriptID, nil
}

// SetRootEvent function
// This will overwrite any events in the script
func (h *ScriptHandler) SetRootEvent(scriptID string, eventID string) (err error) {
	script, err := h.scriptsdb.GetScriptByID(scriptID)
	if err != nil {
		return err
	}
	rootEvent, err := h.eventhandler.eventsdb.GetEventByID(eventID)
	if err != nil {
		return err
	}

	// Set the root eventID
	script.EventIDs = []string{rootEvent.ID}

	err = h.scriptsdb.SaveScriptToDB(script)
	if err != nil {
		return err
	}

	// First we get the events from the root event
	eventlist, err := h.GetDataEvents(rootEvent.ID)
	if err != nil {
		return err
	}
	if len(eventlist) > 0 {
		// While we have events in the list
		for len(eventlist) > 0 {
			// Iterate through each event in the list
			for _, eventinlist := range eventlist {
				h.AddEventToScriptList(eventinlist, script.ID)

				// Find each event in the data fields of the event in the list we are parsing
				foundevents, err := h.GetDataEvents(eventinlist)
				if err != nil {
					return err
				}
				// If we found any events, we add them to the list
				if len(foundevents) > 0 {
					eventlist = append(eventlist, foundevents...)
				}
				// Now we remove the event we just searched from the list
				eventlist = RemoveStringFromSlice(eventlist, eventinlist)
			}
		}
	}
	return nil
}

// AddEventToScriptList function
func (h *ScriptHandler) AddEventToScriptList(eventID string, scriptID string) (err error) {
	script, err := h.scriptsdb.GetScriptByID(scriptID)
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
func (h *ScriptHandler) ExecuteScript(scriptID string, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {
	// First verify the script exists and that the execution paths are valid
	script, err := h.scriptsdb.GetScriptByID(scriptID)
	if err != nil {
		return err
	}

	if len(script.EventIDs) <= 0 {
		return errors.New("script has not been setup yet")
	}

	rootEvent, err := h.eventhandler.eventsdb.GetEventByID(script.EventIDs[0])
	if err != nil {
		return err
	}

	if rootEvent.Watchable {
		err = h.eventhandler.AddEventToWatchList(rootEvent, m.ChannelID)
		if err != nil {
			return err
		}
	}
	return nil
}
