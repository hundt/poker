package user

import (
    "appengine"
    "appengine/datastore"
    //"appengine/user"
    "net/http"
    "poker"
)

type UserState struct {
    Deck poker.Deck
    Position int
    Hand poker.ChineseHand
}

func GetUserState(r *http.Request) (*UserState, error) {
    c := appengine.NewContext(r)
    //u := user.Current(c)
    us := &UserState{Deck: poker.NewShuffledDeck(), Position: 0, Hand: poker.ChineseHand{}}
    _, err := datastore.Put(c, datastore.NewIncompleteKey(c, "UserState", nil), &us)
    if err != nil {
        return nil, err
    }
    return us, nil
}