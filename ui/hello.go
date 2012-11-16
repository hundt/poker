package ui

import (
    "appengine"
    "appengine/channel"
    "appengine/user"
    "fmt"
    "html/template"
    "net/http"
    "poker"
    "strconv"
    //"user"
)

func init() {
    http.HandleFunc("/", pick)
    http.HandleFunc("/compare", compare)
    http.HandleFunc("/create", createGame)
    http.HandleFunc("/game", goToGame)
    http.HandleFunc("/play", play)
}

func createGame(w http.ResponseWriter, r *http.Request) {
    n := r.FormValue("n")
    if n == "" {
        n = "1"
    }
    i, err := strconv.Atoi(n);
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    g := poker.NewGame(i)
    err = g.Save(r);
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    http.Redirect(w, r, "/game?id=" + g.Id(), http.StatusFound)
}

func addWatcher(gs *poker.GameState, uid string) bool {
    for _, watcher := range gs.Watchers {
        if watcher == uid {
            return false
        }
    }
    gs.Watchers = append(gs.Watchers, uid)
    return true
}

func defineNames(w http.ResponseWriter) {
    fmt.Fprint(w, "<script>var names = [];")
    for i := 0; i < 52; i++ {
        fmt.Fprintf(w, "names[%d] = '%s';", i, poker.Card(i).String())
    }
    fmt.Fprint(w, "</script>")
}

func goToGame(w http.ResponseWriter, r *http.Request) {
    g, err := poker.LoadGame(r.FormValue("id"), r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    c := appengine.NewContext(r)
    u := user.Current(c)
    tok, err := channel.Create(c, u.Email + g.Id())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if addWatcher(g, u.Email) {
        err := g.Save(r);
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    }
    json, err := g.ClientState().JSON()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defineNames(w);
    if err := gameTemplate.Execute(w, json); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    fmt.Fprintf(w, "<script>channel('%s')</script>", tok)
}

func play(w http.ResponseWriter, r *http.Request) {
    g, err := poker.LoadGame(r.FormValue("id"), r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    idx, err := strconv.Atoi(r.FormValue("idx"))
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    pos, err := strconv.Atoi(r.FormValue("pos"))
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if pos < poker.Back || pos > poker.Front || idx < 0 || idx >= len(g.Showing) {
        http.Error(w, "Invalid play", http.StatusInternalServerError)
    }
    hand := g.Hands[g.Turn]
    card := g.Showing[idx]
    switch (pos) {
        case poker.Back:
            if hand.Back.Count() >= 5 {
                http.Error(w, "Invalid play", http.StatusInternalServerError)
                return
            }
            g.Hands[g.Turn].Back = hand.Back.Add(card)
        case poker.Middle:
            if hand.Middle.Count() >= 5 {
                http.Error(w, "Invalid play", http.StatusInternalServerError)
                return
            }
            g.Hands[g.Turn].Middle = hand.Middle.Add(card)
        case poker.Front:
            if hand.Front.Count() >= 3 {
                http.Error(w, "Invalid play", http.StatusInternalServerError)
                return
            }
            g.Hands[g.Turn].Front = hand.Front.Add(card)
        default:
            http.Error(w, "Invalid play", http.StatusInternalServerError)
            return
    }
    // Remove the card that was placed
    copy(g.Showing[idx:], g.Showing[idx+1:]) 
    g.Showing = g.Showing[:len(g.Showing)-1]
    if len(g.Showing) == 0 {
        // Someone else's turn
        g.Turn = (g.Turn + 1) % len(g.Hands)
        n := 1
        c := g.Hands[g.Turn].Count()
        if c == 0 {
            n = 5
        } else if (c == 13) {
            n = 0
        }
        g.Showing = g.Deck[g.DeckPos:g.DeckPos + n]
        g.DeckPos = g.DeckPos + n
    }
    if err = g.Save(r); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    json, err := g.ClientState().JSON()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    c := appengine.NewContext(r)
    for _, watcher := range g.Watchers {
        err := channel.Send(c, watcher + g.Id(), json)
        if err != nil {
            c.Errorf("sending Game: %v", err)
        }
    }
    //fmt.Fprint(w, json)
}

func compare(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-type", "text/html; charset=utf-8")
    h1 := poker.NewHand(nil)
    h2 := poker.NewHand(nil)
    for i := 0; i < 52; i++ {
        v := r.FormValue(fmt.Sprintf("%d", i))
        if v == "1" {
            h1 = h1.Add(poker.Card(i))
        } else if v == "2" {
            h2 = h2.Add(poker.Card(i))
        }
    }
    fmt.Fprintf(w, "Hand 1: %s<br>", h1)
    fmt.Fprintf(w, "Hand 2: %s<br>", h2)
    c := h1.Compare(h2)
    if c < 0 {
        fmt.Fprint(w, "Hand 2 wins!")
    } else if c > 0 {
        fmt.Fprint(w, "Hand 1 wins!")
    } else {
        fmt.Fprint(w, "Tie!")
    }
}

func pick(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-type", "text/html; charset=utf-8")
    d := poker.NewOrderedDeck()
    //fmt.Fprint(w, d)
    if err := cardTemplate.Execute(w, d); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

var gameTemplate = template.Must(template.New("game").Parse(gameTemplateHTML))
const gameTemplateHTML = `
<html>
  <body>
  <script type="text/javascript" src="/_ah/channel/jsapi"></script>
  <div id="content">
  <div id="hand0" style="display:none">
  <b>Hand 1:</b> [<span id=hand0royalties>-</span>]<br>
  Back: <span id="hand0back"></span><br>
  Middle: <span id="hand0middle"></span><br>
  Front: <span id="hand0front"></span><br>
  <div id="hand0fault"><b>Fault!</b></div>
  <div id="hand0next">Next Card: <span id="hand0nextCard"></span>
    <input type=button value=Back onclick='play(0)'><input type=button value=Middle onclick='play(1)'><input type=button value=Front onclick='play(2)'>
  </div>
  </div>
  <br>
  <div id="hand1" style="display:none">
  <b>Hand 2:</b> [<span id=hand1royalties>-</span>]<br>
  Back: <span id="hand1back"></span><br>
  Middle: <span id="hand1middle"></span><br>
  Front: <span id="hand1front"></span><br>
  <div id="hand1fault"><b>Fault!</b></div>
  <div id="hand1next">Next Card: <span id="hand1nextCard"></span>
    <input type=button value=Back onclick='play(0)'><input type=button value=Middle onclick='play(1)'><input type=button value=Front onclick='play(2)'>
  </div>
  </div>
  <br>
  <div id="hand2" style="display:none">
  <b>Hand 3:</b> [<span id=hand2royalties>-</span>]<br>
  Back: <span id="hand2back"></span><br>
  Middle: <span id="hand2middle"></span><br>
  Front: <span id="hand2front"></span><br>
  <div id="hand2fault"><b>Fault!</b></div>
  <div id="hand2next">Next Card: <span id="hand2nextCard"></span>
    <input type=button value=Back onclick='play(0)'><input type=button value=Middle onclick='play(1)'><input type=button value=Front onclick='play(2)'>
  </div>
  </div>
  <br>
  <div id="hand3" style="display:none">
  <b>Hand 4:</b> [<span id=hand3royalties>-</span>]<br>
  Back: <span id="hand3back"></span><br>
  Middle: <span id="hand3middle"></span><br>
  Front: <span id="hand3front"></span><br>
  <div id="hand3fault"><b>Fault!</b></div>
  <div id="hand3next">Next Card: <span id="hand3nextCard"></span>
    <input type=button value=Back onclick='play(0)'><input type=button value=Middle onclick='play(1)'><input type=button value=Front onclick='play(2)'>
  </div>
  </div>
  </div>
  <script>
    var hands = [document.getElementById('hand0'),
      document.getElementById('hand1'),
      document.getElementById('hand2'),
      document.getElementById('hand3')];

    var backs = [document.getElementById('hand0back'),
      document.getElementById('hand1back'),
      document.getElementById('hand2back'),
      document.getElementById('hand3back')];
      
    var middles = [document.getElementById('hand0middle'),
      document.getElementById('hand1middle'),
      document.getElementById('hand2middle'),
      document.getElementById('hand3middle')];
      
    var fronts = [document.getElementById('hand0front'),
      document.getElementById('hand1front'),
      document.getElementById('hand2front'),
      document.getElementById('hand3front')];
      
    var nexts = [document.getElementById('hand0next'),
      document.getElementById('hand1next'),
      document.getElementById('hand2next'),
      document.getElementById('hand3next')];
      
    var faults = [document.getElementById('hand0fault'),
      document.getElementById('hand1fault'),
      document.getElementById('hand2fault'),
      document.getElementById('hand3fault')];
      
    var royalties = [document.getElementById('hand0royalties'),
      document.getElementById('hand1royalties'),
      document.getElementById('hand2royalties'),
      document.getElementById('hand3royalties')];
  
    function play(idx, pos) {
       var xhReq = new XMLHttpRequest();
       xhReq.open("GET", "/play?idx=" + idx + "&pos=" + pos + "&id=" + game_id, false);
       xhReq.onreadystatechange = function() {
         if (xhReq.status != 200) {
           alert(xhReq.responseText);
         } else {
           //handle(eval('(' + xhReq.responseText + ')'));
         }
       }
       xhReq.send(null);
    }
    
    function isWinner(i, winners) {
      for (var j = 0; j < winners.length; j++) {
        if (winners[j] == i) return true;
      }
      return false;
    }
    
    function showCards(hand, elt, winner) {
      //alert(hand);
      var html = '';
      if (hand != null) {
        for (var i = 0; i < hand['Royalty']['Cards'].length; i++) {
          html += names[hand['Royalty']['Cards'][i]] + ' ';
        }
        if (hand['Kickers'] != null) {
          for (var i = 0; i < hand['Kickers'].length; i++) {
            html += names[hand['Kickers'][i]] + ' ';
          }
        }
      }
      if (winner) {
        html += " *";
      }
      elt.innerHTML = html;
    }
  
    function showHand(state, i) {
      var hand = state['Hands'][i];
      hands[i].style.display = 'block';
      nexts[i].style.display = 'none';
      faults[i].style.display = 'none';
      royalties[i].childNodes[0].data = state['Royalties'][i];
      if (state['Faults'] && state['Faults'][i]) {
        faults[i].style.display = 'block';
      }
      if (state['Turn'] == i && state['Showing'] && state['Showing'].length != 0) {
        var html = "Dealt:";
        for (var j = 0; j < state['Showing'].length; j++) {
          html += '<br>' + names[state['Showing'][j]];
          html += '<input type=button value=Back onclick="play(' + j + ',0)">';
          html += '<input type=button value=Middle onclick="play(' + j + ',1)">';
          html += '<input type=button value=Front onclick="play(' + j + ',2)">';
        }
        nexts[i].innerHTML = html;
        nexts[i].style.display = 'block';
        //nextCards[i].innerHTML = names[state['Card']]
      }
      showCards(hand['Back'], backs[i], isWinner(i, state['BackWinners']));
      showCards(hand['Middle'], middles[i], isWinner(i, state['MiddleWinners']));
      showCards(hand['Front'], fronts[i], isWinner(i, state['FrontWinners']));
    }
    
    var game_id;
  
    function handle(state) {
      game_id = state['GameId'];
      var hands = state['Hands'];
      //alert(game_id);
      //alert(hands);
      for (var i = 0; i < hands.length; i++) {
        showHand(state, i);
      }
    }
    
    function channel(tok) {
      channel = new goog.appengine.Channel(tok);
      socket = channel.open();
      socket.onopen = function() {};
      socket.onmessage = function(msg) {handle(eval('(' + msg['data'] + ')'))};
      socket.onerror = function(msg) {alert('error: ' + msg['code'] + ' ' + msg['description'])};
      socket.onclose = function() {alert('closed')};;
    }
    handle(eval('(' + {{.}} + ')'))
  </script>
  </body>
</html>
`

var cardTemplate = template.Must(template.New("card").Parse(cardTemplateHTML))

const cardTemplateHTML = `
<html>
  <body>
    <form method=get action=/compare>
    <input type=submit value=Compare> <input type=reset><br>
    {{range .}}
      {{.HTML}} <input type=radio name={{.Id}} value=1>Hand 1&nbsp;&nbsp;&nbsp;<input type=radio name={{.Id}} value=2>Hand 2&nbsp;&nbsp;&nbsp;<input type=radio name={{.Id}} value=0 checked>Neither<br>
    {{end}}
    </form>
  </body>
</html>
`