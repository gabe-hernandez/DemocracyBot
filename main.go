package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"regexp"
	"errors"

	"github.com/bwmarrin/discordgo"
)


const invalidCommand = "I'm afraid I can't do that...(yet)"
const yesVote = "👍"
const noVote = "👎"
const unknownError = "Unknown Error!"

var activeVotes map[string]map[string]string
var usernameToID map[string]*discordgo.User
var defaultVoteTime time.Duration
var defaultVoteTimeStr = "30m"
var threshold = 5
var voters []string
var debug bool

func main() {
	token, exists := os.LookupEnv("DemocracyBotKey")
	_, debug = os.LookupEnv("DemocracyBotDebug")
	if debug {
		defaultVoteTimeStr = "10s"
		threshold = 1
	}
	if !exists {
		fmt.Println("Discord key missing. Please set the DemocracyBotKey environment variable.")
		os.Exit(1)
	}
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	defaultVoteTime, _ = time.ParseDuration(defaultVoteTimeStr)
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

func strInSlice(str string, slice []string) bool {
	for _, v := range slice {
		if strings.Compare(v, str) == 0 {
			return true
		}
	}
	return false
}

func getEmojiID(name string, guild *discordgo.Guild) string {
	for _, v := range guild.Emojis {
		if strings.Compare(v.Name, name) == 0 {
			return v.ID
		}
	}
	return ""
}

func reactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if _, ok := activeVotes[r.MessageID]; ok {
		if r.Emoji.Name == yesVote || r.Emoji.Name == noVote {
			if vote, ok := activeVotes[r.MessageID][r.UserID]; ok {
				s.MessageReactionRemove(r.ChannelID, r.MessageID, vote, r.UserID)
				activeVotes[r.MessageID][r.UserID] = r.Emoji.Name
			} else {
				activeVotes[r.MessageID] = make(map[string]string)
				activeVotes[r.MessageID][r.UserID] = r.Emoji.Name
			}
		}
	}
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
	if debug {
		channel, _ := s.Channel(m.ChannelID)
		if channel.Name != "test" {
			return
		}
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
	case "help":
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Votes last %v long.\nThe required number of votes is %v.", defaultVoteTime, threshold))
	}
}

func vote(s *discordgo.Session, m *discordgo.MessageCreate, commands []string) {
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

func startVote(s *discordgo.Session, m *discordgo.MessageCreate, message string) *discordgo.Message {
	activeVotes[m.ID] = make(map[string]string)
	botMessage, _ := s.ChannelMessageSend(m.ChannelID, message)
	s.MessageReactionAdd(botMessage.ChannelID, botMessage.ID, yesVote)
	s.MessageReactionAdd(botMessage.ChannelID, botMessage.ID, noVote)
	return botMessage
}

func endVote(s *discordgo.Session, m *discordgo.Message) bool {
	result := false
	yesVoters, _ := s.MessageReactions(m.ChannelID, m.ID, yesVote, 100)
	noVoters, _ := s.MessageReactions(m.ChannelID, m.ID, noVote, 100)
	//Don't count the reactions the bot adds
	yesVotes := len(yesVoters) - 1
	noVotes := len(noVoters) - 1
	totalVotes := yesVotes + noVotes
	if yesVotes > noVotes && totalVotes >= threshold {
		s.ChannelMessageSend(m.ChannelID, "The people have spoken!")
		result = true
	} else {
		s.ChannelMessageSend(m.ChannelID, "The vote has failed!")
	}
	activeVotes[m.ID] = make(map[string]string)
	return result
}
func getRoleByName(s *discordgo.Session, guildID string, name string) (string, error) {
	roles, err := s.GuildRoles(guildID)
	for _, role := range roles {
		if role.Name == name {
			return role.ID, nil
		}
	}
	return "", err
}

func roleVote(s *discordgo.Session, m *discordgo.MessageCreate, commands []string) {
	// Currently disabled functionality
	if !debug {
		return
	}
	if len(commands) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Vote role command format is !vote role add/remove user role")
		return
	}
	action := commands[0]
	actionProp := ""
	if action == "add" {
		actionProp = "to"
	} else if action == "remove" {
		actionProp = "from"
	}
	user, err := getUserFromString(s, m.GuildID, commands[1])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "That user doesn't appear to exist!")
		return
	}
	roleName := strings.Join(commands[2:], " ")
	roleID, err := getRoleByName(s, m.GuildID, roleName)
	if err != nil {
		s.ChannelMessage(m.ChannelID, unknownError)
	}
	botM := startVote(s, m, fmt.Sprintf("A vote has started to %v %v %v the role %v! Please react with 👍 or 👎 on this message.", action, user.Username, actionProp, roleName))
	time.Sleep(defaultVoteTime)
	if endVote(s, botM) {
		if action == "add" {
			s.GuildMemberRoleAdd(m.GuildID, user.ID, roleID)
		} else if action == "remove" {
			s.GuildMemberRoleRemove(m.GuildID, user.ID, roleID)
		}
	}

}

func pollVote(s *discordgo.Session, m *discordgo.MessageCreate, commands []string) {
	if len(commands) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Poll command format is !vote poll description")
		return
	}
	botM := startVote(s, m, fmt.Sprintf("A poll has started for %v! Please react with 👍 or 👎 on this message.", strings.Join(commands[:], " ")))
	time.Sleep(defaultVoteTime)
	endVote(s, botM)
}

func getUserFromString(s *discordgo.Session, guildID string, userStr string) (*discordgo.User, error) {
	//When a user is @Named, it gives user string in <@!xxx> format
	if match, _  := regexp.MatchString(`<@\d+>`, userStr); match {
		mem, err := s.GuildMember(guildID, userStr[2:len(userStr)-1])
		if err != nil {
			return nil, err
		}
		return mem.User, nil
	}

	userStr = strings.ToLower(userStr)
	if usernameToID == nil {
		usernameToID = make(map[string]*discordgo.User)
		members, err := s.GuildMembers(guildID, "", 100)

		if err != nil {
			return nil, err
		}

		for _, mem := range members {
			usernameToID[strings.ToLower(mem.User.Username)] = mem.User

			if len(mem.Nick) > 0 {
				usernameToID[strings.ToLower(mem.Nick)] = mem.User
			}
		}
	}

	user, ok := usernameToID[userStr]

	if !ok {
		return nil, errors.New("Username not found")
	}

	return user, nil
}

func nickVote(s *discordgo.Session, m *discordgo.MessageCreate, commands []string) {
	if len(commands) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Vote role command format is !vote role create name")
		return
	}
	
	user, err := getUserFromString(s, m.GuildID, commands[0])
	if  err != nil {
		s.ChannelMessageSend(m.ChannelID, "That user doesn't appear to exist!")
		return
	}
	
	botM := startVote(s, m, fmt.Sprintf("A vote has started to change %v's nickname to %v! Please react with 👍 or 👎 on this message.", user.Username, commands[1]))
	time.Sleep(defaultVoteTime)
	if endVote(s, botM) {
		s.GuildMemberNickname(m.GuildID, user.ID, strings.Join(commands[1:], " "))
	}
}