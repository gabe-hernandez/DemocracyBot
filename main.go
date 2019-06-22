package main

import (
	"errors"
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
var usernameToId map[string]string

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

	if m.Author.Username != "QuantumQuip" && m.Author.Username != "picthebear" && m.Author.Username != "MetalKnux" {
		s.ChannelMessageSend(m.ChannelID, "You have no power here!")
		return
	}

	ID, err := getUserIDfromString(s, m.GuildID, commands[1])

	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		log.Fatal(err)
	}
	s.ChannelMessageSend(m.ChannelID, "The people have spoken!")
	s.GuildMemberNickname(m.GuildID, ID, commands[2])
}

func getUserIDfromString(s *discordgo.Session, guildID string, user string) (string, error) {
	if usernameToId == nil {
		usernameToId = make(map[string]string)
		members, err := s.GuildMembers(guildID, "", 100)

		if err != nil {
			return "", err
		}

		for _, mem := range members {
			usernameToId[mem.User.Username] = mem.User.ID

			if len(mem.Nick) > 0 {
				usernameToId[mem.Nick] = mem.User.ID
			}
		}
	}

	ID, ok := usernameToId[user]

	if !ok {
		return "", errors.New("Username not found")
	}

	return ID, nil
}
