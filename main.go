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
	"time"
	"unicode"

	"github.com/bwmarrin/discordgo"
)

const threshold int = 1

var tokenFile = "bot.key"
var defaultVoteTime, _ = time.ParseDuration("10s")

var usernameToID map[string]string
var voters []string
var protectedUsers map[string]bool

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

	protectedUsers = map[string]bool{"MetalKnuckles": true}

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

func strInSlice(str string, slice []string) bool {
	for _, v := range slice {
		if strings.Compare(v, str) == 0 {
			return true
		}
	}
	return false
}

func inACL(user string) bool {
	switch user {
	case
		"QuantumQuip",
		"picthebear",
		"MetalKnuckles":
		return true
	}
	return false
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content[0] == '!' {
		handleCommand(s, m, m.Content[1:])
	} else {
		return
	}
}

func handleCommand(s *discordgo.Session, m *discordgo.MessageCreate, command string) {
	commands := strings.Split(command, " ")

	switch commands[0] {
	case "vote":
		vote(s, m, commands[1:])
	default:
		s.ChannelMessageSend(m.ChannelID, "I'm afraid I can't do that...(yet)")
	}
}

func vote(s *discordgo.Session, m *discordgo.MessageCreate, commands []string) {
	if len(commands) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Vote command format is !vote action")
		return
	}
	switch commands[0] {
	case "create":
		go beginVote(s, m, commands[1:])
	default:
		s.ChannelMessageSend(m.ChannelID, "I'm afraid I can't do that...(yet)")
	}
}

func beginVote(s *discordgo.Session, m *discordgo.MessageCreate, commands []string) {
	if len(commands) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Vote create command format is !vote create username nickname")
		return
	}
	voters = nil
	ID, err := getUserIDfromString(s, m.GuildID, commands[0])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "That user doesn't appear to exist!")
		return
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Vote to change %v's nickname to %v!", commands[0], commands[1]))
	time.Sleep(defaultVoteTime)
	yesVoters, _ := s.MessageReactions(m.ChannelID, m.ID, "👍", 100)
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("A vote has ended after %v!", defaultVoteTime))

	if threshold <= len(yesVoters) {
		s.ChannelMessageSend(m.ChannelID, "Lol")
		s.GuildMemberNickname(m.GuildID, ID, commands[1])
	} else {
		s.ChannelMessageSend(m.ChannelID, "failed...")
	}
}

func getUserIDfromString(s *discordgo.Session, guildID string, user string) (string, error) {
	//When a user is @Named, it gives user string in <@!xxx> format
	if user[0] == '<' {
		return strings.TrimFunc(user, func(r rune) bool {
			return !unicode.IsNumber(r)
		}), nil
	}

	user = strings.ToLower(user)
	if usernameToID == nil {
		usernameToID = make(map[string]string)
		members, err := s.GuildMembers(guildID, "", 100)

		if err != nil {
			return "", err
		}

		for _, mem := range members {
			usernameToID[strings.ToLower(mem.User.Username)] = mem.User.ID

			if len(mem.Nick) > 0 {
				usernameToID[strings.ToLower(mem.Nick)] = mem.User.ID
			}
		}
	}

	ID, ok := usernameToID[user]

	if !ok {
		return "", errors.New("Username not found")
	}

	return ID, nil
}
