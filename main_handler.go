package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
)

// PrimaryHandler struct
type MainHandler struct {
	db          *DBHandler
	conf        *Config
	dg          *discordgo.Session
	callback    *CallbackHandler
	perm        *PermissionsHandler
	user        *UserHandler
	command     *CommandHandler
	registry    *CommandRegistry
	logchan     chan string
	channel     *ChannelHandler
}

// Init function
func (h *MainHandler) Init() error {
	// DO NOT add anything above this line!!
	// Add our main handler -
	h.dg.AddHandler(h.Read)
	h.registry = h.command.registry

	fmt.Println("Running Startup Setup")
	setup := SetupProcess{db: h.db, conf: h.conf, user: h.user}
	setup.Init(h.dg, h.conf.MainConfig.LobbyChannelID)

	// Add new handlers below this line //
/*
	fmt.Println("Adding Utilities Handler")
	utilities := UtilitiesHandler{db: h.db, conf: h.conf, user: h.user, registry: h.command.registry, logchan: h.logchan, callback: h.callback}
	h.dg.AddHandler(utilities.Read)

	fmt.Println("Adding Tutorial Handler")
	tutorials := TutorialHandler{db: h.db, conf: h.conf, user: h.user, registry: h.command.registry}
	h.dg.AddHandler(tutorials.Read)
*/
	fmt.Println("Adding Notifications Handler")
	notifications := NotificationsHandler{db: h.db, callback: h.callback, conf: h.conf, registry: h.command.registry}
	notifications.Init()
	h.dg.AddHandler(notifications.Read)
	go notifications.CheckNotifications(h.dg)

	// Open a websocket connection to Discord and begin listening.
	fmt.Println("Opening Connection to Discord")
	err := h.dg.Open()
	if err != nil {
		fmt.Println("Error Opening Connection: ", err)
		return err
	}
	fmt.Println("Connection Established")

	err = h.PostInit(h.dg)

	if err != nil {
		fmt.Println("Error during Post-Init")
		return err
	}

	return nil
}

// PostInit function
// Just some quick things to run after our websocket has been setup and opened
func (h *MainHandler) PostInit(dg *discordgo.Session) error {
	fmt.Println("Running Post-Init")

	// Update our default playing status
	fmt.Println("Updating Discord Status")
	err := h.dg.UpdateStatus(0, h.conf.MainConfig.Playing)
	if err != nil {
		fmt.Println("error updating now playing,", err)
		return err
	}

	h.RegisterCommands()

	fmt.Println("Post-Init Complete")
	return nil
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.

// Read function
func (h *MainHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {
	// very important to set this first!
	cp := h.conf.MainConfig.CP

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore bots
	if m.Author.Bot {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding user")
		return
	}

	message := strings.Fields(m.Content)
	if len(message) < 1 {
		fmt.Println(m.Content)
		return
	}

	command := message[0]

	// If the message is "ping" reply with "Pong!"
	if command == cp+"ping" {
		if CheckPermissions("ping", m.ChannelID, &user, s, h.command) {
			s.ChannelMessageSend(m.ChannelID, "Pong!")
			return
		}
	}

	// If the message is "pong" reply with "Ping!"
	if command == cp+"pong" {
		if CheckPermissions("pong", m.ChannelID, &user, s, h.command) {
			s.ChannelMessageSend(m.ChannelID, "Ping!")
			return
		}
	}

	if command == cp+"time" {
		if CheckPermissions("time", m.ChannelID, &user, s, h.command) {
			s.ChannelMessageSend(m.ChannelID, "Current UTC Time: "+time.Now().UTC().Format("2006-01-02 15:04:05"))
			return
		}
	}

	if command == cp+"help" {
		s.ChannelMessageSend(m.ChannelID, "https://github.com/yamamushi/TheAether#table-of-contents")
		return
	}

}

// RegisterCommands function
func (h *MainHandler) RegisterCommands() (err error) {

	h.registry.Register("ping", "Ping command", "ping")
	h.registry.Register("pong", "Pong command", "pong")
	h.registry.Register("time", "Display current UTC time", "time")
	h.registry.Register("tutorial", "Begin the new player tutorial", "tutorial start")

	return nil
}
