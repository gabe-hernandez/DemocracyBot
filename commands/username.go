package commands

import (
	"fmt"
	"time"
	lib "github.com/gabe-hernandez/DemocracyBot/internal"

	"github.com/bwmarrin/discordgo"
)

func changeUsername(s *discordgo.Session, m *discordgo.MessageCreate, commands []string) {
	if len(commands) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Change Name is in form of \"!changename UserName NewUserName\"")
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
