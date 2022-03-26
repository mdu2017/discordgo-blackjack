package cards

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strconv"
)

type Rank struct {
	RankID    int    // Rank identifier
	RankTitle string // Rank title
	RankValue int    // Value of the Rank (if one rank is higher than the other)
	RankCost  int    // Credit cost for this rank
}

// RankList Array holding different ranks in the game
var RankList []Rank

// RankMap maps an id to the rank struct
var RankMap = make(map[int]Rank)

// LoadRankTitles loads all rank titles for the game
func LoadRankTitles() {

	RankList = append(RankList,
		Rank{
			RankID:    1,
			RankTitle: "Top Dog",
			RankValue: 1,
			RankCost:  15000,
		},
		Rank{
			RankID:    2,
			RankTitle: "Almost Top Dog",
			RankValue: 2,
			RankCost:  12000,
		},
		Rank{
			RankID:    3,
			RankTitle: "Pirate",
			RankValue: 3,
			RankCost:  10000,
		},
		Rank{
			RankID:    4,
			RankTitle: "Elite Boss",
			RankValue: 4,
			RankCost:  7500,
		},
		Rank{
			RankID:    5,
			RankTitle: "Boss",
			RankValue: 5,
			RankCost:  7000,
		},
		Rank{
			RankID:    6,
			RankTitle: "Elite Gambler",
			RankValue: 6,
			RankCost:  6500,
		},
		Rank{
			RankID:    7,
			RankTitle: "Gambler",
			RankValue: 7,
			RankCost:  5500,
		},
		Rank{
			RankID:    8,
			RankTitle: "Regular",
			RankValue: 8,
			RankCost:  3000,
		},
		Rank{
			RankID:    9,
			RankTitle: "Elite Player",
			RankValue: 9,
			RankCost:  1500,
		},
		Rank{
			RankID:    10,
			RankTitle: "Player",
			RankValue: 10,
			RankCost:  1000,
		},
		Rank{
			RankID:    11,
			RankTitle: "Starter",
			RankValue: 11,
			RankCost:  0,
		},
	)

	// Create map from array (use starting index at 1 for the map)
	for index, rank := range RankList {
		RankMap[index+1] = rank
	}

}

// CanBuyNextRank Check if player rank is high enough to purchase next one (difference of 1 rank up)
func CanBuyNextRank(session *discordgo.Session, msg *discordgo.MessageCreate, player *Player, rankOption int) bool {

	// If you are a higher rank, cannot purchase a rank below them
	if player.Rank.RankValue <= RankMap[rankOption].RankValue {
		session.ChannelMessageSend(msg.ChannelID, "You have already purchased this rank.")
		return false
	}

	// Check if player has enough credits to purchase rank
	if player.Credits < RankMap[rankOption].RankCost {
		session.ChannelMessageSend(msg.ChannelID, "You don't have enough credits to purchase this rank.")
		return false
	}

	// If player rank is more than 1 rank away, then cannot purchase
	if player.Rank.RankValue-RankMap[rankOption].RankValue > 1 {
		session.ChannelMessageSend(msg.ChannelID, "Rank not high enough to purchase.")
		return false
	}

	return true
}

// PrintRanksInOrder return ranks in order as a string
func PrintRanksInOrder() string {
	rankString := ""

	for i := 0; i < len(RankList); i++ {
		rankString += fmt.Sprintf("%d - %s\n", RankList[i].RankID, RankList[i].RankTitle)
	}

	return rankString
}

// RanksMessageEmbedField returns all ranks as a message embed field with costs
func RanksMessageEmbedField() []*discordgo.MessageEmbedField {

	dgMessageEmbedField := []*discordgo.MessageEmbedField{
		{
			Name:  "1 - Top Dog",
			Value: strconv.Itoa(RankList[0].RankCost),
		},
		{
			Name:  "2 - Almost Top Dog",
			Value: strconv.Itoa(RankList[1].RankCost),
		},
		{
			Name:  "3 - Pirate",
			Value: strconv.Itoa(RankList[2].RankCost),
		},
		{
			Name:  "4 - Elite Boss",
			Value: strconv.Itoa(RankList[3].RankCost),
		},
		{
			Name:  "5 - Boss",
			Value: strconv.Itoa(RankList[4].RankCost),
		},
		{
			Name:  "6 - Elite Gambler",
			Value: strconv.Itoa(RankList[5].RankCost),
		},
		{
			Name:  "7 - Gambler",
			Value: strconv.Itoa(RankList[6].RankCost),
		},
		{
			Name:  "8 - Regular",
			Value: strconv.Itoa(RankList[7].RankCost),
		},
		{
			Name:  "9 - Elite Player",
			Value: strconv.Itoa(RankList[8].RankCost),
		},
		{
			Name:  "10 - Player",
			Value: strconv.Itoa(RankList[9].RankCost),
		},
	}

	return dgMessageEmbedField
}
