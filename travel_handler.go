package main

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"strings"
	"errors"
)

type TravelHandler struct {

	conf       *Config
	registry   *CommandRegistry
	callback   *CallbackHandler
	db         *DBHandler
	perms	   *PermissionsHandler
	room      *RoomsHandler
	user	   *UserHandler

}


// Init TravelHandler
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
		fmt.Println("Error finding user")
		return
	}

	if strings.HasPrefix(m.Content, cp+"travel") {
		if h.registry.CheckPermission("travel", user, s, m) {

			command := strings.Fields(m.Content)

			// Grab our sender ID to verify if this user has permission to use this command
			db := h.db.rawdb.From("Users")
			var user User
			err := db.One("ID", m.Author.ID, &user)
			if err != nil {
				fmt.Println("error retrieving user:" + m.Author.ID)
			}

			if user.CheckRole("moderator") {
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
		s.ChannelMessageSend(user.RoomID, "Error retrieving user: " + err.Error())
		return
	}

	room, err := h.room.rooms.GetRoomByID(user.RoomID)
	if err != nil {
		s.ChannelMessageSend(user.RoomID, "Error retrieving room: " + err.Error())
		return
	}

	useroutput :=  "You travel " + command[1] +  " and arrive at " + room.Name
	useroutput = useroutput + "\n```\n" + room.Description + "\n```"

	s.ChannelMessageSend(user.RoomID, useroutput )

	discorduser, err := s.User(user.ID)
	if err != nil{
		s.ChannelMessageSend(m.ChannelID, "Error retrieving user: " + err.Error())
		return
	}
	leaveoutout := discorduser.Username + " has left traveling " + command[1]

	s.ChannelMessageSend(m.ChannelID, leaveoutout)
	return
}

func (h *TravelHandler) Travel(direction string, s *discordgo.Session, m *discordgo.MessageCreate) (err error){

	user, err := h.user.GetUser(m.Author.ID, s, m.ChannelID)
	if err != nil {
		return err
	}

	room, err := h.room.rooms.GetRoomByID(m.ChannelID)
	if err != nil {
		return err
	}

	toroom := ""
	if direction == "up" {
		toroom = room.UpID
	} else if direction == "down" {
		toroom = room.DownID
	} else if direction == "north" {
		toroom = room.NorthID
	} else if direction == "northeast" {
		toroom = room.NorthEastID
	} else if direction == "east" {
		toroom = room.EastID
	} else if direction == "southeast" {
		toroom = room.SouthEastID
	} else if direction == "south" {
		toroom = room.SouthID
	} else if direction == "southwest" {
		toroom = room.SouthWestID
	} else if direction == "west" {
		toroom = room.WestID
	} else if direction == "northwest" {
		toroom = room.NorthWestID
	} else {
		return errors.New("Unrecognized direction: "+direction)

	}

	if toroom == "" {
		return errors.New("There is nowhere to travel in that direction.")

	}

	targetroom, err := h.room.rooms.GetRoomByID(toroom)
	if err != nil {
		return err
	}

	if len(targetroom.RoleIDs) < 1 {
		return errors.New("Target room is not configured properly: " + toroom )
	}

	if len(room.RoleIDs) < 1 {
		return errors.New("From room is not configured properly: " + toroom )
	}


	if len(targetroom.Items) > 0 {
		// This is where the logic for items validation will go

	}

	guildID, err := getGuildID(s, m.ChannelID)
	if err != nil {
		return err
	}

	roles, err := s.GuildRoles(guildID)
	if err != nil {
		return err
	}

	targetrolename := ""
	targetremoverolename := ""
	for _, role := range roles {
		if role.ID == targetroom.RoleIDs[0] {
			targetrolename = role.Name
		}
		if role.ID == room.RoleIDs[0] {
			targetremoverolename = role.Name
		}
	}

	err = h.perms.AddRoleToUser(targetrolename, user.ID, s, m)
	if err != nil{
		fmt.Println("Error adding role to user")
		return err
	}

	err = h.perms.RemoveRoleFromUser(targetremoverolename, user.ID, s, m)
	if err != nil{
		fmt.Println("Error removing role from user")
		return err
	}

	user, err = h.user.GetUser(m.Author.ID, s, m.ChannelID)
	if err != nil {
		return err
	}

	user.RoomID = targetroom.ID
	db := h.db.rawdb.From("Users")
	err = db.Update(&user)
	if err != nil {
		return errors.New("Error updating user record into database!")
	}

	return nil
}