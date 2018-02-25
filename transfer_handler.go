package main

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strconv"
	"strings"
	"time"
)

// TransferHandler struct
type TransferHandler struct {
	db         *DBHandler
	conf       *Config
	dg         *discordgo.Session
	callback   *CallbackHandler
	perms      *PermissionsHandler
	user       *UserHandler
	command    *CommandHandler
	registry   *CommandRegistry
	channel    *ChannelHandler
	rooms      *RoomsHandler
	transferdb *Transfers
}

// Init function
func (h *TransferHandler) Init() {

	h.transferdb = new(Transfers)
	h.transferdb.db = h.db

	h.RegisterCommands()
}

// RegisterCommands function
func (h *TransferHandler) RegisterCommands() (err error) {

	h.registry.Register("transfer", "Transfer Management", "-")
	h.registry.AddGroup("transfer", "moderator")
	return nil

}

// Read function
func (h *TransferHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.MainConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding usermanager")
		return
	}

	if strings.HasPrefix(m.Content, cp+"transfer") {
		if h.registry.CheckPermission("transfer", user, s, m) {

			command := strings.Fields(m.Content)

			// Grab our sender ID to verify if this usermanager has permission to use this command
			db := h.db.rawdb.From("Users")
			var user User
			err := db.One("ID", m.Author.ID, &user)
			if err != nil {
				fmt.Println("error retrieving usermanager:" + m.Author.ID)
			}

			if user.CheckRole("moderator") {
				h.ParseCommand(command, s, m)
			}
		}
	}
}

// AddTransfer function
func (h *TransferHandler) AddTransfer(userID string, fromChannelID string, toChannelID string, targetGuildID string, fromDirection string) (err error) {

	uuid, err := GetUUID()
	if err != nil {
		return err
	}

	transfer := new(Transfer)

	transfer.ID = uuid
	transfer.TargetChannelID = toChannelID
	transfer.TargetGuildID = targetGuildID
	transfer.FromChannelID = fromChannelID
	transfer.FromDirection = fromDirection
	transfer.UserID = userID

	//fmt.Println(transfer.ID + " " + transfer.UserID + " " + transfer.FromChannelID + " " + transfer.TargetChannelID + " " +
	//				transfer.TargetGuildID + " " + transfer.FromDirection)
	return h.transferdb.SaveTransferToDB(*transfer)
}

// ParseCommand function
func (h *TransferHandler) ParseCommand(input []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	_, payload := SplitPayload(input)

	if len(payload) < 1 {
		s.ChannelMessageSend(m.ChannelID, "transfer requires an argument")
		return
	}

	if payload[0] == "test" {
		if len(payload) < 3 {
			s.ChannelMessageSend(m.ChannelID, "test requires two arguments: <userID> <guildID>")
			return
		}

		fmt.Println("User: " + payload[1] + " Guild: " + payload[2])
		userGuildTest := h.IsUserInGuild(payload[1], payload[2])

		result := "Test Result: " + strconv.FormatBool(userGuildTest)
		s.ChannelMessageSend(m.ChannelID, result)
		return
	}

}

// HandleTransfers function
func (h *TransferHandler) HandleTransfers() {
	// Get all transfers and parse them every 3 minutes
	for true {
		time.Sleep(time.Duration(time.Second * 30))

		transfers, err := h.transferdb.GetAllTransfers()
		if err != nil {
			fmt.Print("Error retrieving transfers db: " + err.Error())
		} else {
			for _, transfer := range transfers {
				time.Sleep(time.Duration(time.Second * 5))

				// Verify the usermanager is actually in the guild before proceeding, otherwise
				// They have not accepted the invite yet and we should skip them for now
				if h.IsUserInGuild(transfer.UserID, transfer.TargetGuildID) {

					// Transfer Channel Roles
					err = h.TransferToChannel(transfer.UserID, transfer.TargetGuildID, transfer.FromChannelID, transfer.TargetChannelID, h.dg)
					if err != nil {
						fmt.Println("Error transferring usermanager: " + err.Error())
					}

					// Remove record from transfers
					h.transferdb.RemoveRoomByID(transfer.ID)

					// Create output for channels
					user, err := h.dg.User(transfer.UserID)
					if err != nil {
						fmt.Println("Error retrieving usermanager: " + err.Error())
					}

					if transfer.FromDirection == "below" || transfer.FromDirection == "above" {
						h.dg.ChannelMessageSend(transfer.TargetChannelID, user.Mention()+" has materialized from "+transfer.FromDirection+".")

					} else {
						h.dg.ChannelMessageSend(transfer.TargetChannelID, user.Mention()+" has materialized from the "+transfer.FromDirection+".")
					}

					h.dg.ChannelMessageSend(transfer.FromChannelID, user.Username+" has dematerialized")

				}
			}
		}
	}
}

// IsUserInGuild function
func (h *TransferHandler) IsUserInGuild(userID string, guildID string) (ispresent bool) {

	_, err := h.dg.GuildMember(guildID, userID)
	if err != nil {
		return false
	}

	return true
}

// TransferToChannel function
func (h *TransferHandler) TransferToChannel(userID string, targetGuildID string, fromChannelID string,
	targetChannelID string, s *discordgo.Session) (err error) {

	// First we remove roles
	fromRoom, err := h.rooms.rooms.GetRoomByID(fromChannelID)
	if err != nil {
		return err
	}

	m := new(discordgo.MessageCreate)
	m.Message = new(discordgo.Message)
	m.Message.ChannelID = fromChannelID

	err = h.perms.RemoveRoleFromUser(fromRoom.TravelRoleID, userID, s, m, true)
	if err != nil {
		return err
	}
	h.rooms.RemoveUserIDFromRoomRecord(userID, fromChannelID)
	if err != nil {
		return errors.New("Error removing usermanager record from room: " + err.Error())
	}

	// Now we add roles
	toroom, err := h.rooms.rooms.GetRoomByID(targetChannelID)
	if err != nil {
		return err
	}

	if len(toroom.AdditionalRoleIDs) < 1 {
		return errors.New("Target room not configured properly")
	}

	m = new(discordgo.MessageCreate)
	m.Message = new(discordgo.Message)
	m.Message.ChannelID = targetChannelID

	err = h.perms.AddRoleToUser(toroom.TravelRoleID, userID, s, m, true)
	if err != nil {
		return err
	}
	h.rooms.AddUserIDToRoomRecord(userID, toroom.ID, toroom.GuildID, s)
	if err != nil {
		return errors.New("Error updating usermanager record into room: " + err.Error())
	}

	// Add registered role here
	err = h.perms.AddRoleToUser("registered", userID, s, m, false)
	if err != nil {
		return err
	}

	user, err := h.user.GetUser(userID, s, m.ChannelID)
	if err != nil {
		return err
	}

	user.RoomID = toroom.ID
	user.GuildID = toroom.GuildID
	db := h.db.rawdb.From("Users")
	err = db.Update(&user)
	if err != nil {
		return errors.New("Error updating usermanager record into database")
	}

	err = h.perms.SyncServerRoles(user.ID, user.RoomID, s)
	if err != nil {
		return err
	}
	return nil
}
