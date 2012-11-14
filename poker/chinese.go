package poker

type ChineseHand struct {
    Back, Middle, Front *Hand
}

func getWinners(hands []*Hand) []int {
    best := make([]int, 0)
    for i, hand := range hands {
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

func Winners(hands []*ChineseHand) (b, m, f []int) {
    return getWinners(selectHands(hands, func(h *ChineseHand)*Hand{return h.Back})),
        getWinners(selectHands(hands, func(h *ChineseHand)*Hand{return h.Middle})),
        getWinners(selectHands(hands, func(h *ChineseHand)*Hand{return h.Front}))
}