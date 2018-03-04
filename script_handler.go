package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

// ScriptHandler struct
type ScriptHandler struct {
	scriptsdb       *ScriptsDB
	db              *DBHandler
	registry        *CommandRegistry
	conf            *Config
	eventhandler    *EventHandler
	eventmessagesdb *EventMessagesDB
	roomshandler    *RoomsHandler
	travelhandler   *TravelHandler
}

// Init function
func (h *ScriptHandler) Init() (err error) {
	fmt.Println("Registering Script Handler Command")
	h.scriptsdb = new(ScriptsDB)
	h.scriptsdb.db = h.db
	h.roomshandler.scripts = h
	h.travelhandler.scripts = h
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
	if command == "save" {
		if len(arguments) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Command 'save' expects an arguments: <scriptName>")
			return
		}
		err := h.SaveScript(arguments[0])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error saving script: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Script saved to disk")
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
	newscript := Script{ID: scriptID, Name: scriptName, CreatorID: userID, Description: description}

	err = h.scriptsdb.SaveScriptToDB(newscript)
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

	//fmt.Println("At executable creation time here is the event list : ")
	//fmt.Println(script.EventIDs)
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
	//fmt.Println("Event list: ")
	//fmt.Println(eventlist)
	if len(eventlist) > 0 {
		// While we have events in the list
		err = h.AddEventToScriptList(rootevent.ID, scriptName)
		if err != nil {
			return err
		}
		for len(eventlist) > 0 {
			// Iterate through each event in the list
			for _, eventinlist := range eventlist {
				h.AddEventToScriptList(eventinlist, script.Name)

				// Find each event in the data fields of the event in the list we are parsing
				//fmt.Println("Event in list: " + eventinlist)
				foundevents, err := h.GetDataEvents(eventinlist)
				if err != nil {
					return err
				}
				//fmt.Println("Found events: " )
				//fmt.Println(foundevents)
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
		err = h.scriptsdb.SaveScriptToDB(script)
		if err != nil {
			return err
		}
	}

	// Now that we have an event list, we want to clone it for our script
	script, err = h.scriptsdb.GetScriptByName(scriptName)
	if err != nil {
		return err
	}

	//fmt.Println("Eventids after list build: " )
	//fmt.Println(script.EventIDs)

	var clonedEventList []string
	for _, eventID := range script.EventIDs {
		event, err := h.eventhandler.eventsdb.GetEventByID(eventID)
		if err != nil {
			return err
		}

		clonedEvent := event

		newEventID := strings.Split(GetUUIDv2(), "-")[0]
		clonedEvent.ID = newEventID
		clonedEvent.IsScriptEvent = true
		clonedEvent.OriginalID = event.ID

		err = h.eventhandler.eventsdb.SaveEventToDB(clonedEvent)
		if err != nil {
			return err
		}
		clonedEventList = append(clonedEventList, clonedEvent.ID)
	}
	//fmt.Println("Script eventIDS Before: ")
	//fmt.Println(script.EventIDs)
	script.EventIDs = clonedEventList
	//fmt.Println("Script eventIDS After: ")
	//fmt.Println(script.EventIDs)
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
	//fmt.Println("Before repair here is the list: ")
	//fmt.Println(script.EventIDs)
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
	//fmt.Println("After repair here is the list: ")
	//fmt.Println(script.EventIDs)
	return nil
}

// SaveScript function
func (h *ScriptHandler) SaveScript(scriptName string) (err error) {
	script, err := h.scriptsdb.GetScriptByName(scriptName)
	if err != nil {
		return err
	}
	if _, err := os.Stat("scripts" + scriptName); err == nil {
		return errors.New("script directory already exists - failing")
	}
	CreateDirIfNotExist("scripts")
	CreateDirIfNotExist("scripts/" + scriptName)

	for _, eventID := range script.EventIDs {
		formattedjson, err := h.eventhandler.EventToJSONString(eventID)
		if err != nil {
			return err
		}

		//fmt.Println(formattedjson)
		err = ioutil.WriteFile("scripts/"+scriptName+"/"+eventID+".event", []byte(formattedjson), 0644)
		if err != nil {
			return err
		}
	}

	jsonscript, err := h.ScriptToJSONString(scriptName)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("scripts/"+scriptName+"/"+scriptName+".script", []byte(jsonscript), 0644)
	if err != nil {
		return err
	}
	return nil
}

// ScriptToJSONString function
func (h *ScriptHandler) ScriptToJSONString(scriptName string) (jsonscript string, err error) {
	script, err := h.scriptsdb.GetScriptByName(scriptName)
	if err != nil {
		return "", err
	}

	jsonscript, err = h.ScriptToJSON(script)
	if err != nil {
		return "", err
	}

	return jsonscript, nil
}

// ScriptToJSON function
func (h *ScriptHandler) ScriptToJSON(script Script) (formatted string, err error) {
	marshalledevent, err := json.Marshal(script)
	if err != nil {
		return "", err
	}
	formatted = string(marshalledevent)
	return formatted, nil
}

// LoadScript function
func (h *ScriptHandler) LoadScript(scriptName string) (err error) {
	_, err = h.scriptsdb.GetScriptByName(scriptName)
	if err == nil {
		return errors.New("script with name " + scriptName + " already exists in database - failing")
	}
	if _, err := os.Stat("scripts/" + scriptName); os.IsNotExist(err) {
		return errors.New("script directory does not exist - failing")
	}

	files, err := ioutil.ReadDir("scripts/" + scriptName)
	if err != nil {
		return errors.New("Error reading directory: " + err.Error())
	}

	for _, file := range files {
		if strings.Contains(file.Name(), ".event") {
			data, err := ioutil.ReadFile("scripts/" + scriptName + "/" + file.Name())
			if err != nil {
				return errors.New("Error reading file: " + file.Name() + " - " + err.Error())
			}

			event, err := h.eventhandler.UnmarshalEvent(data)
			if err != nil {
				return errors.New("Error unpacking file: " + file.Name() + " - " + err.Error())
			}

			err = h.eventhandler.eventsdb.SaveEventToDB(event)
			if err != nil {
				return err
			}
		} else if strings.Contains(file.Name(), ".script") {
			data, err := ioutil.ReadFile(file.Name())
			if err != nil {
				return errors.New("Error reading file: " + file.Name() + " - " + err.Error())
			}

			script, err := h.UnmarshalScript(data)
			if err != nil {
				return errors.New("Error unpacking file: " + file.Name() + " - " + err.Error())
			}

			err = h.scriptsdb.SaveScriptToDB(script)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// UnmarshalScript function
func (h *ScriptHandler) UnmarshalScript(data []byte) (script Script, err error) {
	if err := json.Unmarshal(data, &script); err != nil {
		return script, err
	}
	return script, nil
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
			//fmt.Println("Error retrieiving event by id")
			return err
		}

		for _, channelid := range event.Rooms {
			err = h.eventhandler.DisableEvent(eventid, channelid)
			if err != nil {
				//fmt.Println("Disable event failure")
				return err
			}
			h.eventhandler.UnWatchEvent("", eventid, "")
		}
		err = h.eventhandler.eventsdb.RemoveEventFromDB(event)
		if err != nil {
			//fmt.Println("Remove event failure")
			return err
		}
	}

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
			//fmt.Println("Found event in list, returning nil ")
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
	//fmt.Println("Looking for events in: " +rootEvent.ID)
	if h.eventhandler.parser.HasEventsInData(rootEvent) {
		for _, dataeventid := range rootEvent.Data {
			//fmt.Println("Data eventid: " + dataeventid)
			if dataeventid != "nil" {
				datafieldevent, err := h.eventhandler.eventsdb.GetEventByID(dataeventid)
				if err != nil {
					return eventids, err
				}
				//fmt.Println("Found: " + datafieldevent.ID)
				eventids = append(eventids, datafieldevent.ID)
			}
		}
	}
	defaulteventfield, err := h.eventhandler.eventsdb.GetEventByID(rootEvent.DefaultData)
	if err == nil {
		eventids = append(eventids, defaulteventfield.ID)
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

	//fmt.Println("list is " + strconv.Itoa(len(script.EventIDs)))
	//fmt.Println(script.EventIDs)
	//fmt.Println("Executing: " + script.EventIDs[0])
	rootEvent, err := h.eventhandler.eventsdb.GetEventByID(script.EventIDs[0])
	if err != nil {
		return false, err
	}

	// We need to create a container for our event messages
	eventmessagesID := strings.Split(GetUUIDv2(), "-")[0]
	eventmessage := EventMessageContainer{ID: eventmessagesID, ScriptID: script.ID}
	err = h.eventmessagesdb.SaveEventMessageToDB(eventmessage)
	if err != nil {
		return false, err
	}

	//fmt.Println("parsing event: " + rootEvent.ID)
	if rootEvent.Watchable {
		//fmt.Println("Adding to watchlist: " + rootEvent.ID)
		err = h.eventhandler.AddEventToWatchList(rootEvent, m.ChannelID, eventmessagesID)
		if err != nil {
			return false, err
		}
	} else {
		//fmt.Println("Launching root event: " + rootEvent.ID)
		// If we aren't watching the event, we'd like to get the key value response from it
		h.eventhandler.LaunchChildEvent("RootEvent", rootEvent.ID, eventmessagesID, s, m)
	}

	// Now we page the event message container looking for
	for true {
		time.Sleep(time.Duration(time.Second * 3))
		eventmessage, err = h.eventmessagesdb.GetEventMessageByID(eventmessagesID)
		if err != nil {
			return false, err
		}
		if eventmessage.EventsComplete {
			fmt.Println("Caught events complete")
			break
		}
	}

	// Final cleanup to ensure all of our events are disabled
	for _, eventid := range script.EventIDs {
		event, err := h.eventhandler.eventsdb.GetEventByID(eventid)
		if err != nil {
			return false, err
		}

		h.eventhandler.DisableEvent(event.ID, m.ChannelID)
		h.eventhandler.UnWatchEvent(m.ChannelID, event.ID, eventmessagesID)
		h.eventhandler.eventsdb.SaveEventToDB(event)
		h.eventhandler.eventmessages.TerminateEvents(eventmessagesID)
	}

	eventmessage, err = h.eventmessagesdb.GetEventMessageByID(eventmessagesID)
	if err != nil {
		return false, err
	}
	err = h.eventmessagesdb.RemoveEventMessageByID(eventmessagesID)
	if err != nil {
		return false, err
	}

	if eventmessage.CheckError {
		err = errors.New(eventmessage.ErrorMessage)
	}

	if eventmessage.CheckSuccess {
		if eventmessage.Successful {
			return true, err
		}
		return false, err
	}

	return false, err
}
