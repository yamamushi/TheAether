package main

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
)

// TravelHandler struct
type TravelHandler struct {
	conf     *Config
	registry *CommandRegistry
	callback *CallbackHandler
	db       *DBHandler
	perms    *PermissionsHandler
	room     *RoomsHandler
	user     *UserHandler
	transfer *TransferHandler
	scripts  *ScriptHandler
}

// Init function
func (h *TravelHandler) Init() {
	h.RegisterCommands()

}

// RegisterCommands function
func (h *TravelHandler) RegisterCommands() (err error) {

	h.registry.Register("travel", "Travel in a direction", "up|down|north|northeast|etc")
	h.registry.AddGroup("travel", "player")
	return nil

}

// Read function
func (h *TravelHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.MainConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding usermanager")
		return
	}

	if strings.HasPrefix(m.Content, cp+"travel") {
		if h.registry.CheckPermission("travel", user, s, m) {

			command := strings.Fields(m.Content)

			// Grab our sender ID to verify if this usermanager has permission to use this command
			db := h.db.rawdb.From("Users")
			var user User
			err := db.One("ID", m.Author.ID, &user)
			if err != nil {
				fmt.Println("error retrieving usermanager:" + m.Author.ID)
			}

			if user.CheckRole("player") {
				h.ParseCommand(command, s, m)
			}
		}
	}
}

// ParseCommand function
func (h *TravelHandler) ParseCommand(command []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if len(command) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Expected flag for 'travel' command, see command usage for more info")
		return
	}

	err := h.Travel(command[1], s, m)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	user, err := h.user.GetUser(m.Author.ID, s, m.ChannelID)
	if err != nil {
		s.ChannelMessageSend(user.RoomID, "Error retrieving usermanager: "+err.Error())
		return
	}

	fromroom, err := h.room.rooms.GetRoomByID(user.RoomID)
	if err != nil {
		s.ChannelMessageSend(user.RoomID, "Error retrieving room: "+err.Error())
		return
	}

	travelfrom := ""
	if command[1] == "north" {
		travelfrom = "south"
	} else if command[1] == "northeast" {
		travelfrom = "southwest"
	} else if command[1] == "east" {
		travelfrom = "west"
	} else if command[1] == "southeast" {
		travelfrom = "northwest"
	} else if command[1] == "south" {
		travelfrom = "north"
	} else if command[1] == "southwest" {
		travelfrom = "northeast"
	} else if command[1] == "west" {
		travelfrom = "east"
	} else if command[1] == "northwest" {
		travelfrom = "southeast"
	} else if command[1] == "up" {
		travelfrom = "below"
	} else if command[1] == "down" {
		travelfrom = "above"
	}

	transferroom := Room{}
	if fromroom.GuildTransferInvite != "" {
		transferroom, err = h.room.rooms.GetRoomByID(fromroom.TransferRoomID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error retrieving transfer room: "+err.Error())
			return
		}
	}

	// Notify channel that usermanager has left
	discorduser, err := s.User(user.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieving usermanager: "+err.Error())
		return
	}
	leaveoutout := discorduser.Username + " has left traveling " + command[1]
	s.ChannelMessageSend(m.ChannelID, leaveoutout)

	// If we're not leaving the server, we want to notify the channel that the usermanager has arrived
	if travelfrom == "below" || travelfrom == "above" {
		s.ChannelMessageSend(user.RoomID, discorduser.Mention()+" has arrived from "+travelfrom+".")

	} else {
		s.ChannelMessageSend(user.RoomID, discorduser.Mention()+" has arrived from the "+travelfrom+".")
	}

	time.Sleep(3000)
	// If we're leaving this server, we want to avoid sending an arrival message to the holding channel
	if fromroom.GuildTransferInvite != "" {
		// m.ChannelID because this is the channel we are leaving from
		h.HandleServerTransfer(user, fromroom.ID, fromroom.TransferRoomID, transferroom.GuildID, fromroom, travelfrom, s, m)
		return
	}

	return
}

// HandleServerTransfer function
func (h *TravelHandler) HandleServerTransfer(user User, travelfromID string, transerToID string, targetGuildID string, fromroom Room, fromDirection string,
	s *discordgo.Session, m *discordgo.MessageCreate) {

	// We create a private message to send to the usermanager

	privateInviteMessage := ":satellite: You are now traveling through The Aether, please " +
		"click the invite link below to complete your journey. The materialization process may take a few " +
		"minutes to complete depending on the *Materialization Backlog*: "
	privateInviteMessage = privateInviteMessage + fromroom.GuildTransferInvite

	userprivatechannel, err := s.UserChannelCreate(user.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error creating Aether Link: "+err.Error())
		return
	}

	s.ChannelMessageSend(userprivatechannel.ID, privateInviteMessage)

	channel, err := s.Channel(transerToID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error creating Aether Link: "+err.Error())
		return
	}

	if channel.GuildID != targetGuildID {
		s.ChannelMessageSend(m.ChannelID, "Error creating Aether Link: "+err.Error())
		return
	}

	// We create an notification for the transfer_handler
	err = h.transfer.AddTransfer(user.ID, travelfromID, transerToID, targetGuildID, fromDirection)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error creating Aether Link: "+err.Error())
		return
	}

	return
}

// Travel function
func (h *TravelHandler) Travel(direction string, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {

	user, err := h.user.GetUser(m.Author.ID, s, m.ChannelID)
	if err != nil {
		return err
	}

	fromroom, err := h.room.rooms.GetRoomByID(m.ChannelID)
	if err != nil {
		return err
	}

	toroom := ""
	travelscriptName := ""
	if direction == "up" {
		toroom = fromroom.UpID
		travelscriptName = fromroom.UpScriptName
	} else if direction == "down" {
		toroom = fromroom.DownID
		travelscriptName = fromroom.DownScriptName
	} else if direction == "north" {
		toroom = fromroom.NorthID
		travelscriptName = fromroom.NorthScriptName
	} else if direction == "northeast" {
		toroom = fromroom.NorthEastID
		travelscriptName = fromroom.NorthEastScriptName
	} else if direction == "east" {
		toroom = fromroom.EastID
		travelscriptName = fromroom.EastScriptName
	} else if direction == "southeast" {
		toroom = fromroom.SouthEastID
		travelscriptName = fromroom.SouthEastScriptName
	} else if direction == "south" {
		toroom = fromroom.SouthID
		travelscriptName = fromroom.SouthScriptName
	} else if direction == "southwest" {
		toroom = fromroom.SouthWestID
		travelscriptName = fromroom.SouthWestScriptName
	} else if direction == "west" {
		toroom = fromroom.WestID
		travelscriptName = fromroom.WestScriptName
	} else if direction == "northwest" {
		toroom = fromroom.NorthWestID
		travelscriptName = fromroom.NorthWestScriptName
	} else {
		return errors.New("Unrecognized direction: " + direction)
	}

	if toroom == "" {
		return errors.New("There is nowhere to travel in that direction")
	}

	// Here we run our associated travel script, if one exists for the direction we are traveling
	if travelscriptName != "" {
		err = h.ExecTravelScript(travelscriptName, user, s, m)
		if err != nil {
			return errors.New("Cannot travel " + direction + " - " + err.Error())
		}
	}

	targetroom, err := h.room.rooms.GetRoomByID(toroom)
	if err != nil {
		return err
	}

	if len(targetroom.AdditionalRoleIDs) < 1 {
		return errors.New("Target room is not configured properly: " + toroom)
	}

	if len(fromroom.AdditionalRoleIDs) < 1 {
		return errors.New("From room is not configured properly: " + toroom)
	}

	if len(targetroom.Items) > 0 {
		// This is where the logic for items validation will go
	}

	guildID, err := getGuildID(s, m.ChannelID)
	if err != nil {
		return err
	}

	err = h.perms.AddRoleToUser(targetroom.TravelRoleID, user.ID, s, m, true)
	if err != nil {
		return err
	}

	err = h.perms.RemoveRoleFromUser(fromroom.TravelRoleID, user.ID, s, m, true)
	if err != nil {
		return err
	}

	user, err = h.user.GetUser(m.Author.ID, s, m.ChannelID)
	if err != nil {
		return err
	}

	user.RoomID = targetroom.ID
	user.GuildID = guildID

	db := h.db.rawdb.From("Users")
	err = db.Update(&user)
	if err != nil {
		return errors.New("Error updating user record into database")
	}

	err = h.room.AddUserIDToRoomRecord(user.ID, targetroom.ID, guildID, s)
	if err != nil {
		return errors.New("Error updating user record into room: " + err.Error())
	}

	err = h.room.RemoveUserIDFromRoomRecord(user.ID, fromroom.ID)
	if err != nil {
		return errors.New("Error removing user record from room: " + err.Error())
	}
	return nil
}

// ExecTravelScript function
func (h *TravelHandler) ExecTravelScript(scriptName string, user User, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {

	status, err := h.scripts.ExecuteScript(scriptName, s, m)
	if err != nil {
		return err
	}

	if status == false {
		fmt.Println(err.Error())
		if err == nil {
			return errors.New("")
		}
		return err
	}

	return nil

	// Proof of concept, to be updated later when scripts system is in place!
	/*
		if user.Intelligence < 4 {
			return errors.New("You are not smart enough to travel here yet")
		}
		if user.Strength < 3 {
			return errors.New("You are not strong enough to travel here yet")
		}
		if user.Dexterity < 4 {
			return errors.New("You are not dexterious enough to travel here yet")
		}
		if user.Constitution < 3 {
			return errors.New("You do not have enough constitution to travel here yet")
		}
		if user.Wisdom < 3 {
			return errors.New("You are not wise enough to travel here yet")
		}
		if user.Charisma < 2 {
			return errors.New("You are not charismatic enough to travel here yet")
		}
	*/
}
