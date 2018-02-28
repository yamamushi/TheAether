package main

import (
	"encoding/json"
	"errors"
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"os"
	"strings"
    "fmt"
)

// ActionsHandler struct
type ActionsHandler struct {
	actionsdb *ActionsDB
}

// Init function
func (h *ActionsHandler) Init(actionsdir string) (err error) {

	err = h.LoadActions("./actions")
	if err != nil {
		return err
	}
	return nil
}

// Read function
func (h *ActionsHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

}

// LoadActions function
func (h *ActionsHandler) LoadActions(actionsdir string) (err error) {
	// Read the actions directory if it exists
	if _, err := os.Stat(actionsdir); os.IsNotExist(err) {
		return errors.New("actions directory not found: " + err.Error())
	}

	files, err := ioutil.ReadDir(actionsdir)
	if err != nil {
		return errors.New("Error reading directory: " + err.Error())
	}

	for _, file := range files {
		if strings.Contains(file.Name(), ".action") {
			data, err := ioutil.ReadFile(file.Name())
			if err != nil {
				return errors.New("Error reading file: " + file.Name() + " - " + err.Error())
			}

			action, err := h.UnMarshallAction(data)
			if err != nil {
				return errors.New("Error unpacking file: " + file.Name() + " - " + err.Error())
			}

			_, err = h.actionsdb.GetActionByName(action.Name)
			if err != nil {
				err = h.actionsdb.SaveActionToDB(action)
				if err != nil {
					return errors.New("Error saving action to database: " + action.Name + " - " + err.Error())
				}
				fmt.Print("Loaded action: " + action.Name + " into database!")
			}
		}
	}
	return nil
}

// UnMarshallAction function
func (h *ActionsHandler) UnMarshallAction(data []byte) (action Action, err error) {
	if err := json.Unmarshal(data, &action); err != nil {
		return action, err
	}

	// Generate and assign an ID to this action
	id := strings.Split(GetUUIDv2(), "-")
	action.ID = id[0]
	return action, nil
}
