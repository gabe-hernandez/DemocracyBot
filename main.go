package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var tokenFile = "bot.key"
var members []*Member 
func main() {
	token := readKeyFile()
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	// Cleanly close down the Discord session.
	dg.Close()
}

func readKeyFile() string {
	file, err := os.Open(tokenFile)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	tokenSlice, err := ioutil.ReadAll(file)
	return string(tokenSlice)
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	commands := strings.Split(m.Content, " ")

	if strings.Compare(commands[0], "!vote") != 0 {
		return
	}
	fmt.Println(m.Author.Username)
	temp, err := 
	if err != nil {
		log.Fatal(err)
	}

	for _, mem := range temp {
		fmt.Println(mem.User.ID + " " + mem.User.Username + " " mem.Nick)
	}

}

func getUserIDfromString(s *discordgo.Session, user string) string, error {
	if(members == nil) {
		members = s.GuildMembers(m.GuildID, "", 100)
	}
}
