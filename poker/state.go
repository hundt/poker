package poker

import (
    "appengine"
    "appengine/datastore"
    "bytes"
    "encoding/gob"
    "encoding/json"
    "errors"
    //"fmt"
    "net/http"
)

const (
    Back = iota
    Middle
    Front
)

type Play struct {
    Position int  // back, middle, or front
}

type ClientGameState struct {
    Hands []*ChineseHand
    Showing []Card
    Turn int
    Faults []bool
    Royalties []int
    Payouts [][]float32
    BackWinners, MiddleWinners, FrontWinners []int
    Players []string
    MyTurn bool
    InGame bool
    Started bool
    Finished bool
    GameId string
}

type GameState struct {
    Deck
    DeckPos int
    Hands []*ChineseHand
    Button int
    Turn int
    Showing []Card
    Watchers []string
    Players []string
    PlayerNames []string
    key *datastore.Key
}

func (gs *GameState) ClientState(id string) *ClientGameState {
    cgs := &ClientGameState{
        Hands:gs.Hands,
        GameId:gs.key.Encode(),
        Players:gs.PlayerNames,
        Started:gs.Started(),
        Finished:gs.Finished(),
        MyTurn:len(gs.Players) > 0 && gs.Players[gs.Turn] == id,
        InGame:gs.InGame(id),
    }
    for _, hand := range gs.Hands {
        cgs.Royalties = append(cgs.Royalties, hand.Royalties())
    }
    if len(gs.Showing) != 0 {
        cgs.Showing = gs.Showing
        cgs.Turn = gs.Turn
    } else {
        cgs.Faults = Faults(gs.Hands)
    }
    b, m, f := Winners(gs.Hands, cgs.Faults)
    cgs.BackWinners, cgs.MiddleWinners, cgs.FrontWinners = b, m, f
    return cgs
}

func (gs *GameState) InGame(player string) bool {
    for _, p := range gs.Players {
        if p == player {
            return true
        }
    }
    return false
}

func (gs *GameState) NewHand() {
    gs.Deck = NewShuffledDeck()
    gs.DeckPos = 0
    gs.Hands = make([]*ChineseHand, 0)
    for _, _ = range gs.Players {
        gs.Hands = append(gs.Hands, &ChineseHand{})
    }
    gs.Button = (gs.Button + 1)%len(gs.Hands)
    gs.Turn = gs.Button
    gs.Showing = nil
}

func (gs *GameState) Id() string {
    return gs.key.Encode()
}

func (gs *GameState) Finished() bool {
    return len(gs.Hands) > 0 && gs.Hands[gs.Turn].Count() == 13
}

func (gs *GameState) Sit(id, name string) error {
    if len(gs.Players) == 4 {
        return errors.New("Game is full")
    }
    if gs.Started() {
        return errors.New("Game has already started");
    }
    gs.Players = append(gs.Players, id)
    gs.PlayerNames = append(gs.PlayerNames, name)
    gs.Hands = append(gs.Hands, &ChineseHand{})
    return nil
}

func (gs *GameState) Started() bool {
    return gs.DeckPos > 0
}

func (gs *GameState) NextTurn() {
    if (gs.Started()) {
        gs.Turn = (gs.Turn + 1) % len(gs.Hands)
    }
    n := 1
    c := gs.Hands[gs.Turn].Count()
    if c == 0 {
        n = 5
    } else if (gs.Finished()) {
        n = 0
    }
    gs.Showing = gs.Deck[gs.DeckPos:gs.DeckPos + n]
    gs.DeckPos = gs.DeckPos + n
}

func (gs *GameState) Fix() {
    for _, hand := range gs.Hands {
        hand.Fix()
    }
}

func NewGame(players int) *GameState {
    d := NewShuffledDeck()
    h := make([]*ChineseHand, 0)
    for i := 0; i < players; i++ {
        h = append(h, &ChineseHand{})
    }
    return &GameState{d, 0, h, 0, 0, nil, nil, nil, nil, nil}
}

func (gs *GameState) Bytes() ([]byte, error) {
    var buf bytes.Buffer
    enc := gob.NewEncoder(&buf)
    err := enc.Encode(*gs)
    if err != nil {
        return nil, err
    }
    return buf.Bytes(), nil
}

func (gs *ClientGameState) JSON() (string, error) {
    b, err := json.Marshal(*gs)
    if err != nil {
        return "", err
    }
    return string(b), nil
}

func fromBytes(b []byte) (*GameState, error) {
    buf := bytes.NewBuffer(b)
    dec := gob.NewDecoder(buf)
    var gs *GameState
    if err := dec.Decode(&gs); err != nil {
        return nil, err
    }
    return gs, nil
}

type GameData struct {
    Data []byte
}

func LoadGame(id string, r *http.Request) (*GameState, error) {
    key, err := datastore.DecodeKey(id)
    if err != nil {
        return nil, err
    }
    c := appengine.NewContext(r)
    gd := &GameData{}
    if err = datastore.Get(c, key, gd); err != nil {
        return nil, err;
    }
    gs, err := fromBytes(gd.Data)
    if err != nil {
        return nil, err
    }
    gs.key = key
    return gs, nil
}

func (gs *GameState) Save(r *http.Request) (error) {
    c := appengine.NewContext(r)
    b, err := gs.Bytes()
    if err != nil {
        return err
    }
    if gs.key != nil {
        _, err := datastore.Put(c, gs.key, &GameData{b})
        if err != nil {
            return err
        }
    } else {
        key, err := datastore.Put(c, datastore.NewIncompleteKey(c, "GameState", nil), &GameData{b})
        if err != nil {
            return err
        }
        gs.key = key
    }
    return nil
}