package main

import (

	"fmt"
	"flag"
	"os"
	"log"

	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"syscall"

	"github.com/asdine/storm"
	"github.com/bwmarrin/discordgo"


)


// Variables used for command line parameters
var (
	ConfPath string
)

func init() {
	// Read our command line options
	flag.StringVar(&ConfPath, "c", "aetheral-main.conf", "Path to Config File")
	flag.Parse()

	_, err := os.Stat(ConfPath)
	if err != nil {
		log.Fatal("Config file is missing: ", ConfPath)
		flag.Usage()
		os.Exit(1)
	}
}


func main() {

	fmt.Println("\n\n|| Starting Aetheral ||\n")
	log.SetOutput(ioutil.Discard)

	// Setup our tmp directory
	_, err := os.Stat("tmp")
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir("tmp", os.FileMode(0777))
			if err != nil {
				fmt.Println("Could not make tmp directory! " + err.Error())
				return
			}
		}
	}

	// Verify we can actually read our config file
	conf, err := ReadConfig(ConfPath)
	if err != nil {
		fmt.Println("error reading config file at: ", ConfPath)
		return
	}


	// Create / open our embedded database
	db, err := storm.Open(conf.MainConfig.DBFile)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()


	// Run a quick first time db configuration to verify that it is working properly
	fmt.Println("Checking Database")
	dbhandler := DBHandler{conf: &conf, rawdb: db}
	err = dbhandler.FirstTimeSetup()
	if err != nil {
		log.Fatal(err)
		return
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + conf.MainConfig.Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	defer dg.Close()
	
	if conf.MainConfig.Profiler {
		http.ListenAndServe(":8080", http.DefaultServeMux)
	}


	logchannel := make(chan string)
	logger := Logger{logchan: logchannel}

	// Create a callback handler and add it to our Handler Queue
	fmt.Println("Adding Callback Handler")
	callbackhandler := CallbackHandler{dg: dg, logger: &logger}
	dg.AddHandler(callbackhandler.Read)

	// Create our user handler
	fmt.Println("Adding User Handler")
	userhandler := UserHandler{conf: &conf, db: &dbhandler, logchan: logchannel}
	userhandler.Init()
	dg.AddHandler(userhandler.Read)

	// Create our permissions handler
	fmt.Println("Adding Permissions Handler")
	permissionshandler := PermissionsHandler{dg: dg, conf: &conf, callback: &callbackhandler, db: &dbhandler,
		user: &userhandler, logchan: logchannel}
	dg.AddHandler(permissionshandler.Read)

	// Create our command handler
	fmt.Println("Adding Command Registry Handler")
	commandhandler := CommandHandler{dg: dg, db: &dbhandler, callback: &callbackhandler,
		user: &userhandler, conf: &conf, perm: &permissionshandler, logchan: logchannel}

	// Create our permissions handler
	fmt.Println("Adding Channel Permissions Handler")
	channelhandler := ChannelHandler{db: &dbhandler, conf: &conf, registry: commandhandler.registry,
		user: &userhandler, logchan: logchannel}
	channelhandler.Init()
	dg.AddHandler(channelhandler.Read)

	// Don't forget to initialize the command handler -AFTER- the Channel Handler!
	commandhandler.Init(&channelhandler)
	dg.AddHandler(commandhandler.Read)

	fmt.Println("Initializing Rooms Handler")
	roomshandler := RoomsHandler{callback: &callbackhandler, conf: &conf, db: &dbhandler, perm: &permissionshandler,
	registry: commandhandler.registry, dg: dg, user: &userhandler, ch: &channelhandler}
	dg.AddHandler(roomshandler.Read)
	// No rooms handler init here!

	fmt.Println("Initializing Travel Handler")
	travelhandler := TravelHandler{db: &dbhandler, conf: &conf, registry: commandhandler.registry, perms: &permissionshandler,
	room: &roomshandler, user: &userhandler}
	travelhandler.Init()
	dg.AddHandler(travelhandler.Read)

	// Initalize our Logger
	fmt.Println("Initializing Logger")
	logger.Init(&channelhandler, logchannel, dg)

	// Now we create and initialize our main handler
	fmt.Println("\n|| Initializing Main Handler ||\n")
	handler := MainHandler{db: &dbhandler, conf: &conf, dg: dg, callback: &callbackhandler, perm: &permissionshandler,
		command: &commandhandler, logchan: logchannel, user: &userhandler, channel: &channelhandler, rooms: &roomshandler,
		travel: &travelhandler}
	err = handler.Init()
	if err != nil {
		fmt.Println("error in mainHandler.init", err)
		return
	}
	fmt.Println("\n|| Main Handler Initialized ||\n")


	// Setup Profiler if enabled in config
	if conf.MainConfig.Profiler {
		http.ListenAndServe(":8080", http.DefaultServeMux)
	}


	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

}