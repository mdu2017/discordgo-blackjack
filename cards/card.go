package cards

import (
	"fmt"
	"math/rand"
	"time"
)

// CARDS_IN_DECK Constants for deck
var CARDS_IN_DECK int = 52

// Suit contains a suit and a symbol (suits have no ranking in blackjack)
type Suit struct {
	Name   string
	Symbol string // Can use an icon or picture for later
}

// Card contains a suit and value (Ace of Spades, Three of clubs, etc)
type Card struct {
	Suit  Suit
	Name  string
	Value int
}

// Helper functions below for comparing and checking
func (c *Card) IsFaceCard() bool {
	return c.Name == "Jack" || c.Name == "Queen" || c.Name == "King"
}

func (c *Card) IsAce() bool {
	return c.Name == "Ace"
}

func (c *Card) LessThan(other *Card) bool {
	return c.Value < other.Value
}

func (c *Card) EqualTo(other *Card) bool {
	return c.Value == other.Value
}

func (c *Card) GreaterThan(other *Card) bool {
	return c.Value < other.Value
}

func (c *Card) ToString() string {
	return fmt.Sprintf("%v of %v (%d)", c.Name, c.Suit.Name, c.Value)
}

// Deck interface for a deck (contains a collection of cards)
type Deck struct {
	Cards []Card
	Size  int
}

// CreateDeck creates a new deck of cards (Optional argument for # of decks to use)
func (deck *Deck) CreateDeck(numDecks int) {
	deck.Cards = []Card{}
	deck.Size = CARDS_IN_DECK * numDecks

	// Create 4 types of suits
	suits := []Suit{
		Suit{
			Name:   "Clubs",
			Symbol: "club",
		},
		Suit{
			Name:   "Spades",
			Symbol: "spade",
		},
		Suit{
			Name:   "Hearts",
			Symbol: "heart",
		},
		Suit{
			Name:   "Diamonds",
			Symbol: "diamond",
		},
	}

	// Map of cards with name and values
	nameValues := make(map[string]int)
	nameValues["Two"] = 2
	nameValues["Three"] = 3
	nameValues["Four"] = 4
	nameValues["Five"] = 5
	nameValues["Six"] = 6
	nameValues["Seven"] = 7
	nameValues["Eight"] = 8
	nameValues["Nine"] = 9
	nameValues["Ten"] = 10
	nameValues["Jack"] = 10
	nameValues["Queen"] = 10
	nameValues["King"] = 10
	nameValues["Ace"] = 11

	// Create deck with suits x cards
	for i := 0; i < numDecks; i++ {
		for _, suit := range suits {
			for name, value := range nameValues {
				deck.Cards = append(deck.Cards, Card{
					Suit:  suit,
					Name:  name,
					Value: value,
				})
			}
		}
	}
}

// PrintDeck prints the deck of cards in a linear format
func (deck *Deck) PrintDeck() {
	for _, card := range deck.Cards {
		fmt.Println(card.ToString())
	}
	fmt.Println("Deck size:", deck.Size)
}

// PrintArrayDeck prints deck of cards as an array
func (deck *Deck) PrintArrayDeck() {
	for _, card := range deck.Cards {
		fmt.Print(card.Value, " ")
	}
	fmt.Println()
}

// CardsRemaining current number of cards left in deck
func (deck *Deck) CardsRemaining() int {
	return len(deck.Cards)
}

// Reshuffle re-shuffles the deck
func (deck *Deck) Reshuffle() {
	// Seed random number generator
	rand.Seed(time.Now().Unix())

	// Shuffle function takes in length of array and anonymous swap function
	rand.Shuffle(len(deck.Cards), func(i, j int) {
		deck.Cards[i], deck.Cards[j] = deck.Cards[j], deck.Cards[i]
	})
}

// DrawCard grabs a card from the front of the deck, reduce deck size by 1
func (deck *Deck) DrawCard() Card {

	// If less than 20 cards, then create a new deck
	if deck.CardsRemaining() < 20 {
		deck.CreateDeck(1)
		deck.Reshuffle()
	}

	card := deck.Cards[0]

	deck.Cards = deck.Cards[1:len(deck.Cards)]

	return card
}
