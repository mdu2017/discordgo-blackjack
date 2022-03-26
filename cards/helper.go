package cards

import (
	"strconv"
)

// Constants for emojis
const (
	STOP_SIGN_EMOJI  = "\U0001F6D1" // For quitting the game 🛑
	CHECKBOX_APPROVE = "\U00002705" // For approving double bet ✅
	CHECKBOX_DECLINE = "\U0000274C" // For declining double bet ❌
	TAP_HIT          = "\U0001F446" // Hit in game 👆
	TAP_STAND        = "\U0000270B" // Stand in game ✋
)

// DealStartingHand Deals initial hand to dealer and player (2 cards each)
func DealStartingHand(deck *Deck) ([]Card, []Card) {
	var myHand []Card
	var dealerHand []Card

	card1 := deck.DrawCard()
	card2 := deck.DrawCard()
	card3 := deck.DrawCard()
	card4 := deck.DrawCard()

	myHand = append(myHand, card1, card3)
	dealerHand = append(dealerHand, card2, card4)

	return myHand, dealerHand
}

// IsBlackjack checks if player is dealt a blackjack (Ace + K,Q,J)
func IsBlackjack(hand []Card) bool {
	return (hand[0].IsAce() && hand[1].IsFaceCard()) || (hand[1].IsAce() && hand[0].IsFaceCard())
}

// HandValue returns the numeric value of the hand (two numbers if aces are in play - adjusts so hand doesn't bust)
func HandValue(hand []Card) int {
	var handValue int = 0

	// Scan through to check number of aces in hand
	numAces := 0
	for _, card := range hand {
		if card.Name == "Ace" {
			numAces++
		} else {
			handValue += card.Value
		}
	}

	// Add 11 as value of ace unless it goes over 21
	for i := 0; i < numAces; i++ {
		if (handValue + 11) > 21 {
			handValue++
		} else {
			handValue += 11
		}
	}

	return handValue
}

// IsBust returns if a player's hand is over 21 (bust)
func IsBust(hand []Card) bool {
	return HandValue(hand) > 21
}

// ContainsAce returns true if ace is found in player's hand, otherwise false
func ContainsAce(hand []Card) bool {
	for _, card := range hand {
		if card.Name == "Ace" {
			return true
		}
	}

	return false
}

// PrintHand - prints a short version of player's hand (Ace-2-2, King-10)
func PrintHand(hand []Card) string {
	handString := "Cards in Hand: "

	for index, card := range hand {
		if card.IsFaceCard() || card.IsAce() {
			handString += card.Name
		} else {
			handString += strconv.Itoa(card.Value)
		}

		if index < len(hand)-1 {
			handString += "-"
		}
	}

	return handString
}
