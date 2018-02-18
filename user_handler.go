package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
)

// UserHandler struct
type UserHandler struct {
	conf    *Config
	db      *DBHandler
	cp      string
	logchan chan string
}

// Init function
func (h *UserHandler) Init() {
	h.cp = h.conf.MainConfig.CP
}

// Read function
func (h *UserHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.MainConfig.CP

	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore bots
	if m.Author.Bot {
		return
	}

	message := strings.Fields(m.Content)

	if len(message) < 1 {
		return
	}

	h.CheckUser(m.Author.ID, s, m.ChannelID)

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding user")
		return
	}

	// We use this a bit, this is the author id formatted as a mention
	mention := m.Author.Mention()

	if message[0] == cp+"groups" {
		mentions := m.Mentions

		if len(mentions) == 0 {
			groups, err := h.GetGroups(user.ID, s, m.ChannelID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error retrieving groups: "+err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, h.FormatGroups(groups))
			return
		}

		if !user.CheckRole("moderator") {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command")
			return
		}

		if len(message) == 2 {
			groups, err := h.GetGroups(mentions[0].ID, s, m.ChannelID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error retrieving groups: "+err.Error())
				h.logchan <- "Bot " + mention + " || " + m.Author.Username + " || " + "groups" + "||" + err.Error()
				return
			}
			s.ChannelMessageSend(m.ChannelID, h.FormatGroups(groups))
			return
		}
	}

	if message[0] == cp+"repairuser" {
		if !user.CheckRole("moderator") {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command")
			return
		}

		mentions := m.Mentions

		if len(message) == 2 {
			err := h.RepairUser(mentions[0].ID, s, m.ChannelID, "")
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error repairing user: "+err.Error())
				h.logchan <- "Bot " + mention + " || " + m.Author.Username + " || " + "repairuser" + "||" + err.Error()
				return
			}
			return
		}

		if len(message) == 3 {
			err := h.RepairUser(mentions[0].ID, s, m.ChannelID, message[2])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error repairing user: "+err.Error())
				h.logchan <- "Bot " + mention + " || " + m.Author.Username + " || " + "repairuser" + "||" + err.Error()
				return
			}
			return
		}

	}


	if message[0] == cp+"debuguser"{
		if !user.CheckRole("moderator") {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command")
			return
		}
		mentions := m.Mentions

		if len(message) == 2 {
			err := h.DebugUser(mentions[0].ID, s, m.ChannelID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error repairing user: "+err.Error())
				h.logchan <- "Bot " + mention + " || " + m.Author.Username + " || " + "repairuser" + "||" + err.Error()
				return
			}
			return
		}

	}

	return
}


func (h *UserHandler) AddItem(itemid string, userid string,  s *discordgo.Session, channelID string) (err error) {

	// Make sure user is in the database before we pull it out!
	user, err := h.GetUser(userid, s, channelID)
	if err != nil {
		return err
	}

	user.ItemsMap = append(user.ItemsMap, itemid)
	return nil
}


func (h *UserHandler) RemoveItem(itemid string, userid string,  s *discordgo.Session, channelID string) (err error) {

	// Make sure user is in the database before we pull it out!
	user, err := h.GetUser(userid, s, channelID)
	if err != nil {
		return err
	}

	user.ItemsMap = RemoveStringFromSlice(user.ItemsMap, itemid)
	return nil
}


func (h *UserHandler) RepairUser(userid string, s *discordgo.Session, channelID string, guildID string) (err error) {

	// Make sure user is in the database before we pull it out!
	user, err := h.GetUser(userid, s, channelID)
	if err != nil {
		return err
	}

	if guildID == "" {
		guildID, err = getGuildID(s, channelID)
		if err != nil {
			return err
		}
	}
	user.GuildID = guildID


	db := h.db.rawdb.From("Users")

	err = db.Update(&user)
	if err != nil {
		fmt.Println(":rotating_light: Error updating user record into database!")
		return
	}

	s.ChannelMessageSend(channelID, ":construction: User record repaired!")
	return nil
}

func (h *UserHandler) DebugUser(userid string, s *discordgo.Session, channelID string) (err error) {

	user, err := h.GetUser(userid, s, channelID)
	if err != nil {
		return err
	}

	discordUser, err := s.User(userid)
	if err != nil{
		return err
	}


	userRecord := "```\n"
	userRecord = userRecord + "Username: " + discordUser.Username + "\n"
	userRecord = userRecord + "UserID: " + user.ID + "\n"
	userRecord = userRecord + "GuildID: " + user.GuildID + "\n"
	userRecord = userRecord + "ChannelID: " + user.RoomID + "\n"
	userRecord = userRecord + "Roles: " + "\n"
	userRecord = userRecord + "```\n"

	s.ChannelMessageSend(channelID, "User Record: " + userRecord)

	return nil
}


// GetUser function
func (h *UserHandler) GetUser(userid string, s *discordgo.Session, channelID string) (user User, err error) {

	// Make sure user is in the database before we pull it out!
	h.CheckUser(userid, s, channelID)

	db := h.db.rawdb.From("Users")
	err = db.One("ID", userid, &user)
	if err != nil {
		return user, err
	}

	return user, nil
}

// CheckUser function
func (h *UserHandler) CheckUser(ID string, s *discordgo.Session, channelID string) {

	db := h.db.rawdb.From("Users")

	var u User
	err := db.One("ID", ID, &u)
	if err != nil {
		//fmt.Println("Adding new user to DB: " + ID)

		guildID, err := getGuildID(s, channelID)
		if err != nil {
			fmt.Println("Error retrieving guildID for user: " + err.Error())
			return
		}

		user := User{ID: ID, GuildID: guildID}
		user.Init()

		err = db.Save(&user)
		if err != nil {
			fmt.Println("Error inserting user into Database!")
			return
		}

		return
	}
}




// GetGroups function
func (h *UserHandler) GetGroups(ID string, s *discordgo.Session, channelID string) (groups []string, err error) {

	h.CheckUser(ID, s, channelID)
	user, err := h.GetUser(ID, s, channelID)
	if err != nil {
		return groups, err
	}
	if user.CheckRole("owner") {
		groups = append(groups, "owner")
	}
	if user.CheckRole("admin") {
		groups = append(groups, "admin")
	}
	if user.CheckRole("smoderator") {
		groups = append(groups, "smoderator")
	}
	if user.CheckRole("moderator") {
		groups = append(groups, "moderator")
	}
	if user.CheckRole("builder") {
		groups = append(groups, "builder")
	}
	if user.CheckRole("writer") {
		groups = append(groups, "writer")
	}
	if user.CheckRole("scripter") {
		groups = append(groups, "scripter")
	}
	if user.CheckRole("architect") {
		groups = append(groups, "architect")
	}
	if user.CheckRole("player") {
		groups = append(groups, "player")
	}

	return groups, nil
}

// FormatGroups function
func (h *UserHandler) FormatGroups(groups []string) (formatted string) {
	for i, group := range groups {
		if i == len(groups)-1 {
			formatted = formatted + group
		} else {
			formatted = formatted + group + ", "
		}

	}

	return formatted
}
