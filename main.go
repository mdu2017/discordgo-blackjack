package main

import (
	"bufio"
	"database/sql"
	"discordgo-blackjack/cards"
	"discordgo-blackjack/data"
	"discordgo-blackjack/handler"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/lib/pq"
)

// Game sequence notes

/*
- Deal out initial cards (2 cards to dealer and player)
- Check if either player or dealer has blackjack

- *For discordgo only -- hide 1 of the dealer's cards

- Begin player turn (until player stands or busts)
  - Hit - adds a card from deck to player's hand
  - Stand - finishes player turn - dealer goes next
  - Quit - quits game

- Begin dealer turn (dealer stands at soft 17, or busts)
  - Hit - adds card from deck to dealer's hand
  - Stand - finishes dealer turn
*/

// EmojiList holds all the emoji icons for reactions
var EmojiList map[string]string

// GameDeck - Holds info for game deck and player/dealer hands
var GameDeck cards.Deck
var PlayerHand []cards.Card
var DealerHand []cards.Card
var GameEmbed *discordgo.MessageEmbed

var GameStarted bool = false

// UserProfiles - map of user id to players
// NOTE: this needs to be a pointer to structs so that values in map can be modified
var UserProfiles = make(map[string]*cards.Player)

var BetMultiplier int = 1

var TokenFileName string = "./token.txt"

var DBController *handler.BaseHandler

var BOTID string

func main() {

	// Read token from file and close
	file, err := os.Open(TokenFileName)
	if err != nil {
		log.Fatal("Error opening token file")
	}
	reader := bufio.NewScanner(file)
	reader.Scan()
	token := reader.Text()
	reader.Scan()
	BOTID = reader.Text()
	file.Close()

	bot, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("Error creating discord session")
	}
	// Need to add OnReadyHandler before opening connection, since it loads in the data
	bot.AddHandler(OnReadyHandler)

	// Open a connection
	err = bot.Open()
	if err != nil {
		log.Fatal("Error connecting to discord to discord", err)
	}

	// NOTE: In discordgo, add handlers to listen for events such as creating a message, or on a reaction
	//  -- similar to discord.py on_message, on_reaction_add
	bot.AddHandler(CommandHandler)
	bot.AddHandler(ReactionHandler)

	// Wait until CTRL-C or process is interrupted to stop
	fmt.Println("Bot is now running, press CTRL-C to exit.")
	signalChannel := make(chan os.Signal, 1)

	// Registers a channel to receive notifications of specific signal (executes blockin receive for signals)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-signalChannel

	// Close connection
	bot.Close()
}

// OnReadyHandler This is called when the bot is loaded up and connects
// Used to preload bot data, etc
func OnReadyHandler(session *discordgo.Session, rdy *discordgo.Ready) {

	// Load rank titles
	cards.LoadRankTitles()

	// ============ Database connection ==============
	db := data.OpenDBConnection()
	DBController = handler.NewBaseHandler(db)

	// create table works
	sqlCreateTable := `CREATE TABLE IF NOT EXISTS Player(
		user_id   varchar(20),
		guild_id  varchar(20),
		username  varchar(70),
		credits   int,
		wins      int,
		losses    int,
		user_rank int,

		PRIMARY KEY(user_id, guild_id));
	`

	_, err := DBController.GetDBConn().Exec(sqlCreateTable) // returns (result, error)
	if err != nil {
		fmt.Println("Error creating tables")
		panic(err)
	}

	LoadUserData(session, rdy, DBController.GetDBConn())
}

// Loads user data when bot starts up
func LoadUserData(session *discordgo.Session, rdy *discordgo.Ready, db *sql.DB) {

	// Global emoji list
	EmojiList = make(map[string]string)

	// NOTE: Use session to get the current guild, rdy guild[0] is missing info
	for _, gd := range session.State.Guilds {

		sqlCheckDataExists := `SELECT COUNT(*) FROM Player WHERE guild_id=$1`
		var membersLoaded int

		// Query for row in table, then scan into selected variables
		row := db.QueryRow(sqlCheckDataExists, gd.ID)

		switch err := row.Scan(&membersLoaded); err {
		case sql.ErrNoRows:
			log.Println("No rows returned")
		case nil:

			// If no members from guild is loaded (if 0 -> then create new profiles for each user)
			if membersLoaded == 0 {
				// Leave the after argument as empty to get all guild members
				members, _ := session.GuildMembers(gd.ID, "", 10)

				sqlInsertNewPlayerProfile := `INSERT INTO Player
					(user_id, guild_id, username, credits, wins, losses, user_rank)
					VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING`

				// Load all members into a map for quick access
				for _, member := range members {
					UserProfiles[member.User.ID] = &cards.Player{
						Name:    member.User.Username,
						GuildID: gd.ID,
						Credits: 1000,
						Wins:    0,
						Losses:  0,
						Rank:    cards.RankMap[11],
					}

					// Insert new player data into the database
					_, err := DBController.GetDBConn().Exec(sqlInsertNewPlayerProfile, member.User.ID, gd.ID, member.User.Username,
						1000, 0, 0, 11)
					if err != nil {
						panic(err)
					}
				}
				fmt.Println("New player data created")

			} else {
				// TODO: Query for row of users from db and load them into the map
				sqlGetPlayerRecords := `SELECT user_id, guild_id, username, credits, wins, losses, user_rank 
					FROM Player WHERE guild_id=$1`
				rows, err := db.Query(sqlGetPlayerRecords, gd.ID)
				if err != nil {
					panic(err)
				}
				defer rows.Close() // Rows is a pointer, needs to be closed

				// Go through each row and create player
				for rows.Next() {
					var userid string
					var gid string
					var username string
					var credits int
					var wins int
					var losses int
					var rank int
					err = rows.Scan(&userid, &gid, &username, &credits, &wins, &losses, &rank)
					if err != nil {
						panic(err)
					}
					UserProfiles[userid] = &cards.Player{
						Name:    username,
						GuildID: gid,
						Credits: credits,
						Wins:    wins,
						Losses:  losses,
						Rank:    cards.RankMap[rank],
					}
				}

				// get any error encountered during the iteration
				err = rows.Err()
				if err != nil {
					panic(err)
				}

				fmt.Println("Existing player data loaded")

			}

		default:
			panic(err)
		}

		// Load emoji data
		guild, _ := session.Guild(gd.ID)
		for _, emoji := range guild.Emojis {
			if _, ok := EmojiList[emoji.Name]; ok {
				fmt.Println("Emoji name already exists in map")
			} else {
				// NOTE: NEED to use APINAME() to pass to MessageReactionAdd for the emojiID (discord format is emojiName:emojiID)
				EmojiList[emoji.Name] = emoji.APIName()
			}
		}

		// NOTE: Manually loading in emojis to debug issue with unresponsive reactions
		EmojiList["CHECKBOX_APPROVE"] = "✅"
		EmojiList["CHECKBOX_DECLINE"] = "❌"
		EmojiList["TAP_HIT"] = "👆"
		EmojiList["TAP_STAND"] = "✋"
		EmojiList["STOP"] = "🛑"

	}
	fmt.Println("Member data loaded successfully")
	fmt.Println("Emoji Data loaded successfully")

}

// CommandHandler Handles commands when user types in a message
func CommandHandler(session *discordgo.Session, msg *discordgo.MessageCreate) {
	// Ignore self-messages from bot
	if msg.Author.ID == session.State.User.ID {
		return
	}

	// Set prefix as !game for commands
	if !strings.HasPrefix(msg.Content, "!game") {
		return
	}

	// Get arguments from message
	args := strings.Split(msg.Content, " ")
	args = args[1:]

	// Make sure command has a valid argument
	if len(args) < 1 {
		log.Println("Make sure to enter a command. Type !game help to see a list of commands.")
		return
	}

	// First argument should be the command name
	commandName := args[0]

	// If argument is length 2, grab the 2nd argument as shop option
	var shopChoice int
	if len(args) == 2 {
		shopChoice, _ = strconv.Atoi(args[1])
	}

	// Start game, show help, etc
	switch commandName {
	case "blackjack":

		// Set initial bet multiplier to 1
		BetMultiplier = 1.0

		// Initialize card deck
		GameDeck.CreateDeck(6)
		GameDeck.Reshuffle()
		GameStarted = true

		// Deal starting cards
		PlayerHand, DealerHand = cards.DealStartingHand(&GameDeck)

		session.ChannelMessageSend(msg.ChannelID, "Welcome to Blackjack. Dealing out initial hands.")

		GameEmbed = &discordgo.MessageEmbed{
			Title: "Blackjack Table",
			Description: fmt.Sprintf("Two initial cards have been dealt to the dealer and player. Click %s to double bet, or %s to continue",
				cards.CHECKBOX_APPROVE, cards.CHECKBOX_DECLINE),
			Color:  0,
			Footer: nil,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name: fmt.Sprintf("%s %s", UserProfiles[msg.Author.ID].Rank.RankTitle, msg.Author.Username),
					Value: fmt.Sprintf("User %v's turn. React with %s to hit, %s to stand, or %s tp quit",
						msg.Author.Username, cards.TAP_HIT, cards.TAP_STAND, cards.STOP_SIGN_EMOJI),
				},
				{
					Name:  "Your current Hand",
					Value: fmt.Sprintf("%v\n", cards.PrintHand(PlayerHand)),
				},
			},
		}

		embed, _ := session.ChannelMessageSendEmbed(msg.ChannelID, GameEmbed)

		// Wait for user to double bet or continue
		// TODO: Currently bugged for wait on reaction
		session.MessageReactionAdd(msg.ChannelID, embed.ID, EmojiList["CHECKBOX_APPROVE"])
		session.MessageReactionAdd(msg.ChannelID, embed.ID, EmojiList["CHECKBOX_DECLINE"])
		// _ = <-waitForReaction(session)

		time.AfterFunc(time.Second*3, func() {

			// NOTE: This is the only message remove function that I found that will work for the moment (RemoveAll doesn't work)
			session.MessageReactionRemove(msg.ChannelID, embed.ID, "✅", session.State.User.ID)
			session.MessageReactionRemove(msg.ChannelID, embed.ID, "❌", session.State.User.ID)

			session.MessageReactionAdd(msg.ChannelID, embed.ID, cards.TAP_HIT)
			session.MessageReactionAdd(msg.ChannelID, embed.ID, cards.TAP_STAND)
			session.MessageReactionAdd(msg.ChannelID, embed.ID, cards.STOP_SIGN_EMOJI)
		})

		// If player gets an immediate blackjack then end game
		if cards.IsBlackjack(PlayerHand) {
			session.ChannelMessageSendEmbed(msg.ChannelID, &discordgo.MessageEmbed{
				Title:       "Blackjack! You win!",
				Description: "You earned 750 credits",
				Color:       0,
			})
			GameStarted = false

			// Add 750 credits to player for a blackjack
			// END GAME
			if player, ok := UserProfiles[msg.Author.ID]; ok {
				player.Credits += 750
				player.Wins += 1
			}

			return
		} else if cards.IsBlackjack(DealerHand) {
			session.ChannelMessageSendEmbed(msg.ChannelID, &discordgo.MessageEmbed{
				Title:       "Dealer Blackjack! You lost!",
				Description: "You lost 750 credits",
				Color:       0,
			})
			GameStarted = false

			// Deduct 750 credits from player if they lose
			if player, ok := UserProfiles[msg.Author.ID]; ok {
				player.Credits -= 750
				player.Losses += 1

				// Make sure you can't have negative credits
				if player.Credits < 0 {
					player.Credits = 0
				}
			}

			return
		}

		// TODO: the rest of game logic is in CommandHandler when the player reacts to the given table reactions

	case "wallet":
		DisplayPlayerCredits(session, msg)
	case "save":
		SavePlayerData(session, msg, data.DBConn)
	case "stats":
		DisplayPlayerStats(session, msg)
	case "shop":
		DisplayGameShop(session, msg)
	case "ranks":
		DisplayRanks(session, msg)
	case "buy":
		PurchaseRankTitle(session, msg, shopChoice)

	case "help":
		helpEmbed := &discordgo.MessageEmbed{
			URL:         "",
			Type:        "",
			Title:       "List of Commands",
			Description: "List of Commands available",
			Timestamp:   "",
			Color:       0,
			Footer:      nil,
			Image:       nil,
			Thumbnail:   nil,
			Video:       nil,
			Provider:    nil,
			Author:      nil,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "!game help",
					Value: "Display a list of commands",
				},
				{
					Name:  "!game blackjack",
					Value: "Starts a game of blackjack with the CPU",
				},
				{
					Name:  "!game wallet",
					Value: "Shows how many credits you have.",
				},
				{
					Name:  "!game stats",
					Value: "Displays your win-loss record and rank",
				},
				{
					Name:  "!game shop",
					Value: "Displays a list of titles you can purchase",
				},
			},
		}
		_, err := session.ChannelMessageSendEmbed(msg.ChannelID, helpEmbed)
		if err != nil {
			fmt.Println("Error showing help")
			return
		}

	default:
		session.ChannelMessageSend(msg.ChannelID, "Type \"!game help\" for a list of game commands")
	}
}

// ReactionHandler Handles events when a user reacts to a message
func ReactionHandler(session *discordgo.Session, reaction *discordgo.MessageReactionAdd) {

	// Make sure bot doesn't respond to its own reaction
	if reaction.UserID == BOTID {
		return
	}

	// In cases with multiple discord bots, ignore reactions on messages that the bot doesn't send out
	channelMsg, _ := session.ChannelMessage(reaction.ChannelID, reaction.MessageID)
	if channelMsg.Author.ID != BOTID {
		return
	}

	if !GameStarted {
		return
	}

	switch reaction.Emoji.Name {
	case "👆":
		HitReactionHandler(session, reaction)

		// NOTE: Need to find a way to remove user's reaction once clicked on
		// session.MessageReactionRemove(reaction.ChannelID, reaction.MessageID, EmojiList["TAP_HIT"], session.State.User.ID)

	case "✋":
		session.ChannelMessageSend(reaction.ChannelID, "You stand. Dealer's turn...")
		// session.MessageReactionRemove(reaction.ChannelID, reaction.MessageID, "\U0000270B", session.State.User.ID)

		// Dealer's turn
		for {
			dealerHandValue := cards.HandValue(DealerHand)
			fmt.Println("Dealer hand value:", dealerHandValue)

			// ======= Player wins, dealer busts ============
			if dealerHandValue > 21 {
				session.ChannelMessageSend(reaction.ChannelID, fmt.Sprintf("Dealer BUST! Player wins! You get %d credits", 300*BetMultiplier))
				GameStarted = false

				// Add 300 credits to player if they win
				// END GAME
				if player, ok := UserProfiles[reaction.UserID]; ok {
					player.Credits += 300 * BetMultiplier
					player.Wins += 1
				}

				return
			} else if dealerHandValue >= 17 { // Stands at soft 17 (any hand with an ace in it -> A-6, A-3-3, A-4-2)
				session.ChannelMessageSend(reaction.ChannelID, "Dealer stands...")
				GameStarted = false
				break // Make sure to break so dealer doesn't draw another card after standing
			}

			// Draw another card and add to hand if dealer hasn't reached 17
			drawnCard := GameDeck.DrawCard()
			DealerHand = append(DealerHand, drawnCard)
		}

		// Check game win scenarios
		// Player busts -> dealer automatically wins
		// Player doesn't bust, dealer does -> player wins
		// Neither bust -> hand value is compared
		// Tie game

		session.ChannelMessageSend(reaction.ChannelID, fmt.Sprintf("Your hand: %v\nDealer hand: %v\n",
			cards.HandValue(PlayerHand), cards.HandValue(DealerHand)))

		if cards.HandValue(PlayerHand) == cards.HandValue(DealerHand) {
			session.ChannelMessageSend(reaction.ChannelID, "Tie game! Push.")
			GameStarted = false
		} else if cards.HandValue(PlayerHand) < cards.HandValue(DealerHand) {

			fmt.Println("Player Hand: ", cards.PrintHand(PlayerHand), cards.HandValue(PlayerHand))
			fmt.Println("Dealer Hand: ", cards.PrintHand(DealerHand), cards.HandValue(DealerHand))
			session.ChannelMessageSend(reaction.ChannelID, fmt.Sprintf("Dealer Wins! You lost %d credits!", 300*BetMultiplier))

			// END GAME
			if player, ok := UserProfiles[reaction.UserID]; ok {
				player.Credits -= 300 * BetMultiplier
				player.Losses += 1

				// Make sure you can't have negative credits
				if player.Credits < 0 {
					player.Credits = 0
				}
			}

			GameStarted = false

		} else {
			// END GAME
			session.ChannelMessageSend(reaction.ChannelID, fmt.Sprintf("Player Wins! You get %d credits!", 300*BetMultiplier))

			if player, ok := UserProfiles[reaction.UserID]; ok {
				player.Credits += 300 * BetMultiplier
				player.Wins += 1
			}

			GameStarted = false
		}

	case "✅":
		session.ChannelMessageSend(reaction.ChannelID, "Doubling rewards/losses -- 2x Credits")
		// session.MessageReactionRemove(reaction.ChannelID, reaction.MessageID, "✅", currUser.ID)
		BetMultiplier = 2
		GameStarted = true
	case "❌":
		session.ChannelMessageSend(reaction.ChannelID, "Continue with game...")
		// session.MessageReactionRemove(reaction.ChannelID, reaction.MessageID, "❌", currUser.ID)
		GameStarted = true
	case "eight":
		log.Println("Start game with 4 decks")

		//For unicode emojis, just place the actual emoji for the name
	case "🛑":
		session.ChannelMessageSend(reaction.ChannelID, "Quitting game.")
		GameStarted = false
	}
}

// Waits for a reaction and adds a handler to the current session. Returns a channel with the reaction in it.
// func waitForReaction(session *discordgo.Session) chan *discordgo.MessageReactionAdd {
// 	channel := make(chan *discordgo.MessageReactionAdd)
// 	session.AddHandlerOnce(func(_ *discordgo.Session, rxn *discordgo.MessageReactionAdd) {
// 		channel <- rxn
// 	})
// 	return channel
// }

// HitReactionHandler Handles logic when players hit for another card
func HitReactionHandler(session *discordgo.Session, reaction *discordgo.MessageReactionAdd) {

	drawnCard := GameDeck.DrawCard()
	PlayerHand = append(PlayerHand, drawnCard)

	// Update player hand and embed
	GameEmbed.Fields[1].Value = fmt.Sprintf("%v", cards.PrintHand(PlayerHand))
	session.ChannelMessageEditEmbed(reaction.ChannelID, reaction.MessageID, GameEmbed)

	// Check if player has busted
	// END GAME
	if cards.IsBust(PlayerHand) {
		session.ChannelMessageSend(reaction.ChannelID, "You BUST! Dealer wins")
		GameStarted = false

		if user, ok := UserProfiles[reaction.UserID]; ok {
			user.Credits -= 300 * BetMultiplier

			// Make sure you can't have negative credits
			if user.Credits < 0 {
				user.Credits = 0
			}
		}

		return
	}
}

// SavePlayerData Saves a user profile - updates in database
func SavePlayerData(session *discordgo.Session, msg *discordgo.MessageCreate, db *sql.DB) {
	currentPlayerID := msg.Author.ID

	// If player is found in profiles, update the current status
	if currentPlayer, ok := UserProfiles[currentPlayerID]; ok {

		sqlSavePlayerData := `UPDATE Player SET credits=$2, wins=$3, losses=$4, user_rank=$5 WHERE user_id=$1;`
		_, err := DBController.GetDBConn().Exec(sqlSavePlayerData, currentPlayerID, currentPlayer.Credits,
			currentPlayer.Wins, currentPlayer.Losses, currentPlayer.Rank.RankID)
		if err != nil {
			panic(err)
		}

		session.ChannelMessageSend(msg.ChannelID, "Your data has been saved.")
	}

}

// Display player stats
func DisplayPlayerStats(session *discordgo.Session, msg *discordgo.MessageCreate) {

	playerID := msg.Author.ID
	statsEmbed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Stats for %v", msg.Author.Username),
		Description: "Blackjack Records",
		Color:       0,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Wins",
				Value: strconv.Itoa(UserProfiles[playerID].Wins),
			},
			{
				Name:  "Losses",
				Value: strconv.Itoa(UserProfiles[playerID].Losses),
			},
			{
				Name:  "Rank",
				Value: UserProfiles[playerID].Rank.RankTitle,
			},
		},
	}

	_, err := session.ChannelMessageSendEmbed(msg.ChannelID, statsEmbed)
	if err != nil {
		fmt.Println("Error showing stats embed")
		return
	}
}

// Display player credits
func DisplayPlayerCredits(session *discordgo.Session, msg *discordgo.MessageCreate) {

	playerID := msg.Author.ID
	creditsEmbed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Stats for %v", msg.Author.Username),
		Description: "Wallet",
		Color:       0,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Credits",
				Value: strconv.Itoa(UserProfiles[playerID].Credits),
			},
		},
	}

	_, err := session.ChannelMessageSendEmbed(msg.ChannelID, creditsEmbed)
	if err != nil {
		fmt.Println("Error showing creditsEmbed")
		return
	}
}

// Display game shop
func DisplayGameShop(session *discordgo.Session, msg *discordgo.MessageCreate) {
	playerID := msg.Author.ID
	shopEmbed := &discordgo.MessageEmbed{
		Title: "Blackjack Bazaar",
		Description: fmt.Sprintf("Type \"!game buy <X> for the corresponding title to purchase.\n"+
			" e.g !game buy 4 to purchase Title 4\n"+
			"**Your Wallet Total - %s credits**", strconv.Itoa(UserProfiles[playerID].Credits)),
		Color:  0,
		Fields: cards.RanksMessageEmbedField(),
	}

	_, err := session.ChannelMessageSendEmbed(msg.ChannelID, shopEmbed)
	if err != nil {
		fmt.Println("Error showing creditsEmbed")
		return
	}
}

// DisplayRanks Display all ranks
func DisplayRanks(session *discordgo.Session, msg *discordgo.MessageCreate) {
	rankEmbed := &discordgo.MessageEmbed{
		Title:       "List of Ranks",
		Description: cards.PrintRanksInOrder(),
		Color:       0,
	}

	_, err := session.ChannelMessageSendEmbed(msg.ChannelID, rankEmbed)
	if err != nil {
		fmt.Println("Error showing creditsEmbed")
		return
	}
}

// PurchaseRankTitle
func PurchaseRankTitle(session *discordgo.Session, msg *discordgo.MessageCreate, shopChoice int) {

	if player, ok := UserProfiles[msg.Author.ID]; ok {
		if cards.CanBuyNextRank(session, msg, player, shopChoice) {
			player.Credits -= cards.RankMap[shopChoice].RankCost
			player.Rank = cards.RankMap[shopChoice]
			session.ChannelMessageSend(msg.ChannelID,
				fmt.Sprintf("You have purchased the next rank: %s", player.Rank.RankTitle))

			SavePlayerData(session, msg, data.DBConn)
		}
	}
}
