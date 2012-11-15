package poker

type ChineseHand struct {
    Back, Middle, Front *Hand
}

func (ch *ChineseHand) Count() int {
    return ch.Back.Count() + ch.Middle.Count() + ch.Front.Count()
}

func (ch *ChineseHand) Fault() bool {
    return ch.Back.Compare(ch.Middle) < 0 || ch.Middle.Compare(ch.Front) < 0
}

func (ch *ChineseHand) Fix() {
    ch.Back = ch.Back.Fix()
    ch.Middle = ch.Middle.Fix()
    ch.Front = ch.Front.Fix()
}

func (ch *ChineseHand) Royalties() int {
    r := 0
    if (ch.Count() == 13 && ch.Fault()) {
        return 0;
    }
    if (ch.Back != nil) {
        switch (ch.Back.Royalty.Rank) {
            case AceLowStraight:
                r += 2
            case Straight:
                r += 2
            case Flush:
                r += 4
            case FullHouse:
                r += 6
            case Quads:
                r += 8
            case AceLowStraightFlush:
                r += 10
            case StraightFlush:
                r += 10
            case Royal:
                r += 20
        }
    }
    if (ch.Middle != nil) {
        switch (ch.Middle.Royalty.Rank) {
            case AceLowStraight:
                r += 4
            case Straight:
                r += 4
            case Flush:
                r += 8
            case FullHouse:
                r += 12
            case Quads:
                r += 16
            case AceLowStraightFlush:
                r += 20
            case StraightFlush:
                r += 20
            case Royal:
                r += 40
        }
    }
    if (ch.Front != nil) {
        switch (ch.Front.Royalty.Rank) {
            case Pair:
                if cr := int(ch.Front.Royalty.Cards[0].Rank()) - 3; cr > 0 {
                    r += cr;
                }
            case Trips:
                r += int(ch.Front.Royalty.Cards[0].Rank()) + 10;
        }
    }
    return r
}

func Faults(hands []*ChineseHand) []bool {
    result := make([]bool, 0)
    for _, hand := range hands {
        result = append(result, hand.Fault())
    }
    return result
}

func getWinners(hands []*Hand, faults []bool) []int {
    best := make([]int, 0)
    for i, hand := range hands {
        if faults != nil && faults[i] {
            continue;
        }
        if len(best) == 0 {
            // First hand
            best = append(best, i)
            continue
        }
        c := hand.Compare(hands[best[0]])
        if c == 0 {
            best = append(best, i)
        } else if c > 0 {
            best = []int{i}
        }
    }
    return best
}

func selectHands(hands []*ChineseHand, s func(*ChineseHand)*Hand) []*Hand {
    result := make([]*Hand, 0)
    for _, hand := range hands {
        result = append(result, s(hand))
    }
    return result
}

func Winners(hands []*ChineseHand, faults []bool) (b, m, f []int) {
    return getWinners(selectHands(hands, func(h *ChineseHand)*Hand{return h.Back}), faults),
        getWinners(selectHands(hands, func(h *ChineseHand)*Hand{return h.Middle}), faults),
        getWinners(selectHands(hands, func(h *ChineseHand)*Hand{return h.Front}), faults)
}