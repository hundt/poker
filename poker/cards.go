package poker

import (
    "fmt"
    crand "crypto/rand"
    "html/template"
    "math/big"
    "math/rand"
    "time"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

type Card int
type Suit int
type Rank int
type Deck []Card

const (
    HighCard = iota
    Pair
    TwoPair
    Trips
    AceLowStraight
    Straight
    Flush
    FullHouse
    Quads
    AceLowStraightFlush
    StraightFlush
    Royal
)

type Royalty struct {
    Rank int
    Cards []Card
}

func (r Royalty) String() string {
    rs := ""
    switch r.Rank {
        case HighCard:
            rs = "High Card"
        case Pair:
            rs = "Pair"
        case TwoPair:
            rs = "Two Pair"
        case Trips:
            rs = "Trips"
        case AceLowStraight:
            rs = "Straight"
        case Straight:
            rs = "Straight"
        case Flush:
            rs = "Flush"
        case FullHouse:
            rs = "Full House"
        case Quads:
            rs = "Quads"
        case AceLowStraightFlush:
            rs = "Straight Flush"
        case StraightFlush:
            rs = "Straight Flush"
        case Royal:
            rs = "Royal Flush"
    }
    return fmt.Sprintf("%s (%v)", rs, r.Cards)
}

func NewOrderedDeck() Deck {
    d := Deck(make([]Card, 52))
    for i, _ := range d {
        d[i] = Card(i)
    }
    return d
}

func nextInt(n int) int {
    i, err := crand.Int(crand.Reader, big.NewInt(int64(n)))
    if err != nil {
        return r.Intn(n)
    }
    return int(i.Int64())
}

func NewShuffledDeck() Deck {
    d := NewOrderedDeck()
    for i, _ := range d {
        j := i + nextInt(52 - i)
        temp := d[i]
        d[i] = d[j]
        d[j] = temp
    }
    return d
}

func (c Card) Suit() Suit {
    return Suit(c / 13)
}

func (c Card) String() string {
    return fmt.Sprintf("%s%s", c.Rank(), c.Suit())
}

func (c Card) HTML() template.HTML {
    return template.HTML(fmt.Sprintf("%s%s", c.Rank(), c.Suit()))
}

func (c Card) Rank() Rank {
    return Rank(int(c) % 13)
}

func (c Card) Id() int {
    return int(c)
}

func (r Rank) String() string {
    switch r {
        case 9:
            return "J"
        case 10:
            return "Q"
        case 11:
            return "K"
        case 12:
            return "A"
    }
    return fmt.Sprintf("%d", r + 2)
}

func (s Suit) String() string {
    switch s {
        case 0:
            return "\u2660"
        case 1:
            return "\u2663"
        case 2:
            return "<font color=red>\u2665</font>"
        case 3:
            return "<font color=red>\u2666</font>"
    }
    return "ERROR"
}