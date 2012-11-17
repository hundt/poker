package poker

import (
    "fmt"
    "sort"
)

type Hand struct {
    Royalty Royalty
    Kickers []Card
}

func (h *Hand) Count() int {
    if (h == nil) {
        return 0
    }
    return len(h.Kickers) + len(h.Royalty.Cards)
}

func (h *Hand) String() string {
    return fmt.Sprintf("%s Kickers: %v", h.Royalty, h.Kickers)
}

func (h *Hand) Fix() *Hand {
    if (h == nil) {
        return nil
    }
    var nh *Hand;
    for _, card := range h.Royalty.Cards {
        nh = nh.Add(card)
    }
    for _, card := range h.Kickers {
        nh = nh.Add(card)
    }
    return nh
}

func (h *Hand) Add(c Card) *Hand {
    if (h == nil) {
        return NewHand([]Card{c})
    }
    cards := h.Royalty.Cards
    for _, card := range h.Kickers {
        cards = append(cards, card)
    }
    found := false
    for _, other := range cards {
        if other == c {
            found = true
        }
    }
    if !found {
        cards = append(cards, c)
    }
    return NewHand(cards)
}

type CardList []Card

func (cards CardList) Len() int {
    return len(cards)
}

func (cards CardList) Less(i, j int) bool {
    return cards[i].Rank() < cards[j].Rank()
}

func (cards CardList) Swap(i, j int) {
    t := cards[i]
    cards[i] = cards[j]
    cards[j] = t
}

func flushHand(cards []Card) *Hand {
    if len(cards) <= 4 {
        return nil
    }
    suit := cards[0].Suit()
    for _, card := range cards {
        if card.Suit() != suit {
            return nil
        }
    }
    return &Hand{Royalty: Royalty{Flush, cards}, Kickers: nil}
}

func aceLowStraightHand(cards []Card) *Hand {
    if len(cards) == 5 &&
        cards[0].Rank() == 0 &&
        cards[1].Rank() == 1 &&
        cards[2].Rank() == 2 &&
        cards[3].Rank() == 3 &&
        cards[4].Rank() == 12 {
        return &Hand{Royalty: Royalty{AceLowStraight, cards}, Kickers: nil}
    }
    return nil
}

func straightHand(cards []Card) *Hand {
    if len(cards) <= 4 {
        return nil
    }
    r := cards[0].Rank() - 1
    for _, card := range cards {
        if card.Rank() != r + 1 {
            return aceLowStraightHand(cards)
        }
        r++
    }
    return &Hand{Royalty: Royalty{Straight, cards}, Kickers: nil}
}

func filter(cards []Card, rank Rank, invert bool) []Card {
    result := make([]Card, 0)
    for _, card := range cards {
        match := card.Rank() == rank
        if invert {
            match = !match
        }
        if match {
            result = append(result, card)
        }
    }
    return result
}

func quadsHand(cards []Card) *Hand {
    if len(cards) <= 3 {
        return nil
    }
    if q := filter(cards, cards[0].Rank(), false); len(q) == 4 {
        return &Hand{
            Royalty: Royalty{Quads, q},
            Kickers: filter(cards, cards[0].Rank(), true)}
    }
    if q := filter(cards, cards[1].Rank(), false); len(q) == 4 {
        return &Hand{
            Royalty: Royalty{Quads, q},
            Kickers: filter(cards, cards[1].Rank(), true)}
    }
    return nil
}

func maybeFullHouse(cards []Card, tripsRank Rank) *Hand {
    kickers := filter(cards, tripsRank, true)
    if h := pairsHand(kickers); h != nil {
        return &Hand{
            Royalty: Royalty{FullHouse, cards},
            Kickers: h.Kickers}
    }
    return &Hand{
        Royalty: Royalty{Trips, filter(cards, tripsRank, false)},
        Kickers: kickers}
}

func tripsHand(cards []Card) *Hand {
    if len(cards) <= 2 {
        return nil
    }
    if q := filter(cards, cards[0].Rank(), false); len(q) == 3 {
        return maybeFullHouse(cards, cards[0].Rank())
    }
    if q := filter(cards, cards[1].Rank(), false); len(q) == 3 {
        return maybeFullHouse(cards, cards[1].Rank())
    }
    if q := filter(cards, cards[2].Rank(), false); len(q) == 3 {
        return maybeFullHouse(cards, cards[2].Rank())
    }
    return nil
}

func pairsHand(cards []Card) *Hand {
    if len(cards) <= 1 {
        return nil
    }
    for i, _ := range cards {
        if i > len(cards) - 2 {
            break
        }
        if ((cards[i].Rank() == cards[i + 1].Rank()) &&
            (i + 1 == len(cards) -1 || cards[i + 2].Rank() != cards[i].Rank())) {
            kickers := filter(cards, cards[i].Rank(), true)
            cards := []Card{cards[i], cards[i + 1]}
            r := Pair
            if tp := pairsHand(kickers); tp != nil {
                for _, card := range tp.Royalty.Cards {
                    cards = append(cards, card)
                }
                kickers = tp.Kickers
                r = TwoPair
            }
            return &Hand{
                Royalty: Royalty{r, cards},
                Kickers: kickers}
        }
    }
    return nil
}

func straightOrFlushHand(cards []Card) *Hand {
    hand := flushHand(cards)
    if hand != nil {
        if straight := straightHand(hand.Royalty.Cards); straight != nil {
            if straight.Royalty.Rank == AceLowStraight {
                return &Hand{Royalty: Royalty{AceLowStraightFlush, hand.Royalty.Cards}, Kickers: hand.Kickers}
            }
            if cards[0] == 8 {  // ace high
                return &Hand{Royalty: Royalty{Royal, hand.Royalty.Cards}, Kickers: hand.Kickers}
            }
            return &Hand{Royalty: Royalty{StraightFlush, hand.Royalty.Cards}, Kickers: hand.Kickers}
        }
        return hand
    }
    return straightHand(cards)
}

func NewHand(cards []Card) *Hand {
    if len(cards) > 5 || len(cards) < 1 {
        return nil
    }
    sort.Sort(CardList(cards))
    if hand := straightOrFlushHand(cards); hand != nil {
        return hand
    }
    if hand := quadsHand(cards); hand != nil {
        return hand
    }
    if hand := tripsHand(cards); hand != nil {
        return hand
    }
    if hand := pairsHand(cards); hand != nil {
        return hand
    }
    return &Hand{
        Royalty: Royalty{HighCard, []Card{cards[len(cards)-1]}},
        Kickers: cards[0:len(cards)-1]}
}

func min(x, y int) int {
    if x < y {
        return x
    }
    return y
}

type Hands []*Hand

func (h Hands) Less(i, j int) bool {
    return h[i].Compare(h[j]) < 0
}

func (h Hands) Swap(i, j int) {
    temp := h[i]
    h[i] = h[j]
    h[j] = temp
}

func (h Hands) Len() int {
    return len(h)
}

func (h *Hand) Compare(other *Hand) int {
    if h == nil {
        if other == nil {
            return 0
        }
        return -1
    }
    if other == nil {
        return 1
    }
    d := h.Royalty.Rank - other.Royalty.Rank
    if d != 0 {
        return d
    }
    for i, _ := range h.Royalty.Cards {
        idx := len(h.Royalty.Cards) - i - 1
        d = int(h.Royalty.Cards[idx].Rank() - other.Royalty.Cards[idx].Rank())
        if d != 0 {
            return d
        }
    }
    nk := min(len(h.Kickers), len(other.Kickers))
    for i := 0; i < nk; i++ {
        d = int(h.Kickers[len(h.Kickers) - i - 1].Rank() - other.Kickers[len(other.Kickers) - i - 1].Rank())
        if d != 0 {
            return d
        }
    }
    return len(h.Kickers) - len(other.Kickers)
}