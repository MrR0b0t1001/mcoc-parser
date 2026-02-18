package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	mcocParser "github.com/MrR0b0t1001/mcoc-parser/parser"
	"github.com/MrR0b0t1001/mcoc-parser/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

const discordMsgLimit = 1900

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("warning: assuming default configuration. .env unreadable: %v", err)
	}

	// Retrieve the discord bot token
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN not provided")
	}
	// Retrieve the channel ID the bot needs to look for incoming commands
	targetChannelID := os.Getenv("DISCORD_CHANNEL_ID")
	if targetChannelID == "" {
		log.Fatal("DISCORD_CHANNEL_ID not provided")
	}

	// Initiate discord bot session
	sess, err := discordgo.New(
		"Bot " + token,
	)
	if err != nil {
		log.Fatal("bot session failed to initiate %w", err)
	}

	// IMPORTANT: if you want to read m.Content, you must enable Message Content Intent in the Discord portal
	// and include it here.
	sess.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// ignore if author nil or if message sent by bot
		if m.Author == nil || m.Author.Bot {
			return
		}

		// channel the bot needs to look at
		if m.ChannelID != targetChannelID {
			return
		}

		// Retrieve champID from command
		champID, ok := utils.ParseBotCommand(m.Content)
		if !ok {
			return
		}

		// Check to see if champ exists in map
		champ, exists := utils.ChampionList[champID]
		if !exists {
			_, _ = s.ChannelMessageSend(
				m.ChannelID,
				"Unknown champion ID. Try writing /x -> where x U (1, 314)",
			)
			return
		}

		title, html, err := mcocParser.FetchChampionHTML(champ)
		if err != nil {
			log.Printf("error in fetching champion info for: %v", champ)
		}

		champDetails, err := mcocParser.ParseChampion(title, html)
		if err != nil {
			log.Println("oh well")
		}

		result := mcocParser.FormatForDiscord(champDetails)

		if err := utils.SendBatched(s, m.ChannelID, result, discordMsgLimit); err != nil {
			log.Printf("failed sending batched msg: %v", err)
		}
	})

	if err := sess.Open(); err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	log.Println("bot is running...")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
