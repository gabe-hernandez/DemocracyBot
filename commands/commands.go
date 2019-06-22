package commands

import (
	"fmt"
	"strings"
	"time"
	lib "github.com/gabe-hernandez/DemocracyBot/internal"

	"github.com/bwmarrin/discordgo"
)

func HandleCommand(s *discordgo.Session, m *discordgo.MessageCreate, command string) {
	commands := strings.Split(command, " ")
	if !lib.IsAdmin(m.Author.Username) {
		s.ChannelMessageSend(m.ChannelID, "You have no power here!")
		return
	}
	switch strings.ToLower(commands[0]) {
	case "vote":
		vote(s, m, commands[1:])
	case "changename":
		changeUserName(s, m, commands[1:])
	case "test":
		s.ChannelMessageSend(m.ChannelID, "Testing asyncronous calls(go routines)...")
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
		go createVote(s, m, commands[1:])
	default:
		s.ChannelMessageSend(m.ChannelID, "I'm afraid I can't do that...(yet)")
	}
}

func createVote(s *discordgo.Session, m *discordgo.MessageCreate, commands []string) {
	if len(commands) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Vote create command format is !vote create username nickname")
		return
	}
	ID, err := lib.GetUserIDfromString(s, m.GuildID, commands[0])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "That user doesn't appear to exist!")
		return
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("A vote has started to change %v's nickname to %v!", commands[0], commands[1]))
	s.MessageReactionAdd(m.ChannelID, m.ID, "👍")
	s.MessageReactionAdd(m.ChannelID, m.ID, "👎")
	time.Sleep(lib.DefaultVoteTime)
	yesVoters, _ := s.MessageReactions(m.ChannelID, m.ID, "👍", 100)
	//noVoters, _ := s.MessageReactions(m.ChannelID, m.ID, "👎", 100)
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("A vote has ended after %v!", lib.DefaultVoteTime))
	if lib.Threshold <= len(yesVoters) {
		s.ChannelMessageSend(m.ChannelID, "The people have spoken!")
		s.GuildMemberNickname(m.GuildID, ID, commands[1])
	} else {
		s.ChannelMessageSend(m.ChannelID, "The vote has failed!")
	}
}
