package poker

import (
    "appengine"
    "appengine/datastore"
    "bytes"
    "encoding/gob"
    "encoding/json"
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
    GameId string
}

type GameState struct {
    Deck
    DeckPos int
    Hands []*ChineseHand
    Turn int
    Showing []Card
    Watchers []string
    key *datastore.Key
}

func (gs *GameState) Id() string {
    return gs.key.Encode()
}

func (gs *GameState) Fix() {
    for _, hand := range gs.Hands {
        hand.Fix()
    }
}

func (gs *GameState) ClientState() *ClientGameState {
    cgs := &ClientGameState{
        Hands:gs.Hands,
        GameId:gs.key.Encode(),
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

func NewGame(players int) *GameState {
    d := NewShuffledDeck()
    h := make([]*ChineseHand, 0)
    for i := 0; i < players; i++ {
        h = append(h, &ChineseHand{})
    }
    return &GameState{d, 5, h, 0, d[0:5], nil, nil}
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