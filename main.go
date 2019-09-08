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

	"github.com/bwmarrin/discordgo"
)

const threshold = 1
const invalidCommand = "I'm afraid I can't do that...(yet)"
const yesVote = "👍"
const noVote = "👎"

var tokenFile = "bot.key"
var activeVotes map[string]map[string]string
var usernameToId map[string]string
var defaultVoteTime, _ = time.ParseDuration("20s")
var voters []string

func main() {
	token, exists := os.LookupEnv("DemocracyBotKey")
	if !exists {
		fmt.Println("Discord key missing. Please set the DemcoracyBotKey environment variable.")
		os.Exit(1)
	}
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	activeVotes = make(map[string]map[string]string)
	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)
	dg.AddHandler(reactionAdd)

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
		"MetalKnuckles",
		"Baeson":
		return true
	}
	return false
}

func getEmojiId(name string, guild *discordgo.Guild) string {
	for _, v := range guild.Emojis {
		if strings.Compare(v.Name, name) == 0 {
			return v.ID
		}
	}
	return ""
}

func reactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if _, ok := activeVotes[r.MessageID]; ok {
		if strings.Compare(r.Emoji.Name, yesVote) == 0 || strings.Compare(r.Emoji.Name, noVote) == 0 {
			if vote, ok := activeVotes[r.MessageID][r.UserID]; ok {
				s.ChannelMessageSend(r.ChannelID, "Debug: User already voted!")
				s.MessageReactionRemove(r.ChannelID, r.MessageID, vote, r.UserID)
				activeVotes[r.MessageID][r.UserID] = r.Emoji.Name
			} else {
				activeVotes[r.MessageID] = make(map[string]string)
				activeVotes[r.MessageID][r.UserID] = r.Emoji.Name
				s.ChannelMessageSend(r.ChannelID, "Debug: First vote!")
			}
		}
	} else {
		s.ChannelMessageSend(r.ChannelID, "Debug: There is no active vote of this message!")
	}
	return
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	//g, _ := s.Guild(m.GuildID)
	//getEmojiId("thumbsup", g)
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
	if !inACL(m.Author.Username) {
		s.ChannelMessageSend(m.ChannelID, "You have no power here!")
		return
	}
	switch commands[0] {
	case "vote":
		Vote(s, m, commands[1:])
	case "help":
		s.ChannelMessageSend(m.ChannelID, "<:Reverse:497131062653353996>")
	default:
		s.ChannelMessageSend(m.ChannelID, "I'm afraid I can't do that...(yet)")
	}
}

func Vote(s *discordgo.Session, m *discordgo.MessageCreate, commands []string) {
	if len(commands) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Vote command format is !vote action")
		return
	}
	switch commands[0] {
	case "nick":
		go nickVote(s, m, commands[1:])
	case "poll":
		go pollVote(s, m, commands[1:])
	case "role":
		go roleVote(s, m, commands[1:])
	default:
		s.ChannelMessageSend(m.ChannelID, "I'm afraid I can't do that...(yet)")
	}
}

func startVote(s *discordgo.Session, m *discordgo.MessageCreate, message string) {
	activeVotes[m.ID] = make(map[string]string)
	s.ChannelMessageSend(m.ChannelID, message)
}

func endVote(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	result := false
	yesVoters, _ := s.MessageReactions(m.ChannelID, m.ID, "👍", 100)
	noVoters, _ := s.MessageReactions(m.ChannelID, m.ID, "👎", 100)
	yesVotes := len(yesVoters)
	noVotes := len(noVoters)
	totalVotes := yesVotes + noVotes
	if yesVotes > noVotes && totalVotes >= threshold {
		s.ChannelMessageSend(m.ChannelID, "The people have spoken!")
		result = true
	} else {
		s.ChannelMessageSend(m.ChannelID, "The vote has failed!")
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("A vote has ended after %v!", defaultVoteTime))
	activeVotes[m.ID] = make(map[string]string)
	return result
}

func roleVote(s *discordgo.Session, m *discordgo.MessageCreate, commands []string) {
	if len(commands) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Vote role command format is !vote role create name")
		return
	}
	ID, err := getUserIDfromString(s, m.GuildID, commands[0])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "That user doesn't appear to exist!")
		return
	}
	startVote(s, m, fmt.Sprintf("A vote has started to change %v's nickname to %v!", commands[0], commands[1]))
	time.Sleep(defaultVoteTime)
	if endVote(s, m) {
		s.GuildMemberNickname(m.GuildID, ID, commands[1])
	}

}

func pollVote(s *discordgo.Session, m *discordgo.MessageCreate, commands []string) {
	if len(commands) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Poll command format is !vote poll description")
		return
	}
	startVote(s, m, fmt.Sprintf("A poll has started for %v! Please react with 👍 or 👎 on the above message.", strings.Join(commands[:], " ")))
	time.Sleep(defaultVoteTime)
	endVote(s, m)
}

func nickVote(s *discordgo.Session, m *discordgo.MessageCreate, commands []string) {
	if len(commands) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Vote role command format is !vote role create name")
		return
	}
	ID, err := getUserIDfromString(s, m.GuildID, commands[0])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "That user doesn't appear to exist!")
		return
	}
	startVote(s, m, fmt.Sprintf("A vote has started to change %v's nickname to %v! Please react with 👍 or 👎 on the above message.", commands[0], commands[1]))
	time.Sleep(defaultVoteTime)
	if endVote(s, m) {
		s.GuildMemberNickname(m.GuildID, ID, commands[1])
	}
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
