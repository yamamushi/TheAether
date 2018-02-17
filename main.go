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

	fmt.Println(conf.MainConfig.Token)

	if conf.MainConfig.Profiler {
		http.ListenAndServe(":8080", http.DefaultServeMux)
	}

	// Open a websocket connection to Discord and begin listening.
	fmt.Println("Opening Connection to Discord")
	err = dg.Open()
	if err != nil {
		fmt.Println("Error Opening Connection: ", err)
		return
	}
	fmt.Println("Connection Established")


	fmt.Println("Updating Discord Status")
	err = dg.UpdateStatus(0, conf.MainConfig.Playing)
	if err != nil {
		fmt.Println("error updating now playing,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

}