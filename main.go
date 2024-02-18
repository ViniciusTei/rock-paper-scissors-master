package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Choice int

const (
	ROCK Choice = iota + 1
	PAPER
	SCISSORS
)

type User struct {
	Nick   string
	Choice Choice
	Conn   *websocket.Conn
}

type Game struct {
	Id      int
	Player1 *User
	Player2 *User
}

type WSMessage struct {
	Choice  string `json:"choice"`
	Headers struct {
		HXRequest     string `json:"HX-Request"`
		HXTrigger     string `json:"HX-Trigger"`
		HXTriggerName string `json:"HX-Trigger-Name"`
		HXTarget      string `json:"HX-Target"`
		HXCurrentURL  string `json:"HX-Current-URL"`
	} `json:"HEADERS"`
}

var Games = []Game{}

func (d Choice) String() string {
	return [...]string{"rock", "paper", "scissors"}[d-1]
}

func createUser(nickname string) *User {
	user := User{
		Nick:   nickname,
		Choice: 0,
		Conn:   nil,
	}

	return &user
}

func (u *User) Connect(conn *websocket.Conn) error {
	if u == nil {
		return errors.New("Tried to connect player that does not exists")
	}

	u.Conn = conn
	return nil
}

func (u *User) SetChoice(choice string) error {
	if choice == "rock" {
		u.Choice = ROCK
		return nil
	}

	if choice == "paper" {
		u.Choice = PAPER
		return nil
	}

	if choice == "scissors" {
		u.Choice = SCISSORS
		return nil
	}

	return errors.New("Choice not defined")
}

func (u User) HasChoice() bool {
	if u.Choice != 0 {
		return true
	}

	return false
}

func (u *User) SendTemplate(file string, data map[string]string) error {
	var htmlTemplate bytes.Buffer
	tmpl := template.Must(template.ParseFiles(file))
	if errExec := tmpl.Execute(&htmlTemplate, data); errExec != nil {
		log.Println("Error executing template", errExec)
		return errExec
	}

	err := u.Conn.WriteMessage(websocket.TextMessage, []byte(htmlTemplate.String()))
	if err != nil {
		return errors.New(err.Error())
	}
	return nil
}

func (u *User) SendMessage(file string, block string, data map[string]string) error {
	var htmlTemplate bytes.Buffer
	tmpl := template.Must(template.ParseFiles(file))
	if errExec := tmpl.ExecuteTemplate(&htmlTemplate, block, data); errExec != nil {
		log.Println("Error executing template", errExec)
		return errExec
	}

	err := u.Conn.WriteMessage(websocket.TextMessage, []byte(htmlTemplate.String()))
	if err != nil {
		return errors.New(err.Error())
	}
	return nil
}

func (u User) SendResult(result string, opponent *User) {
	if result == "draw" {
		if errPlayer := u.SendTemplate("templates/draw.tmpl", map[string]string{
			"choice":   u.Choice.String(),
			"opponent": opponent.Nick,
			"house":    opponent.Choice.String(),
		}); errPlayer != nil {
			return
		}

		return
	}

	if result == "win" {
		if errPlayer := u.SendTemplate("templates/win.tmpl", map[string]string{
			"choice":   u.Choice.String(),
			"opponent": opponent.Nick,
			"house":    opponent.Choice.String(),
		}); errPlayer != nil {
			return
		}

		return
	}

	if result == "lose" {
		if errPlayer := u.SendTemplate("templates/lose.tmpl", map[string]string{
			"choice":   u.Choice.String(),
			"opponent": opponent.Nick,
			"house":    opponent.Choice.String(),
		}); errPlayer != nil {
			return
		}

		return
	}

}

func (g *Game) connectPlayer2(player *User) {
	g.Player2 = player
}

func createNewGame(player *User) Game {
	game := Game{
		Id:      len(Games),
		Player1: player,
	}

	return game
}

func addNewGame(player *User) Game {
	game := createNewGame(player)
	Games = append(Games, game)
	return game
}

func findGame(id int) *Game {
	if id < 0 || id > len(Games) {
		log.Fatal("Game out of bounds")
	}

	return &Games[id]
}

func (g Game) opponentPlayer(username string) *User {
	if g.Player1 == nil || g.Player1.Conn == nil || g.Player2 == nil || g.Player2.Conn == nil {
		return nil
	}

	if g.Player1.Nick == username {
		return g.Player2
	}

	return g.Player1
}

func (g Game) connectedPlayer(username string) *User {
	if g.Player1.Nick == username {
		return g.Player1
	}
	return g.Player2
}

func (g Game) isPlayerConnected(username string) bool {
	if g.Player1 != nil && g.Player1.Nick == username && g.Player1.Conn != nil {
		return true
	}

	if g.Player2 != nil && g.Player2.Nick == username && g.Player2.Conn != nil {
		return true
	}

	return false
}

func (g Game) handlePlayerMessage(player *User) {
	opponent := g.opponentPlayer(player.Nick)

	if player.HasChoice() && opponent != nil && opponent.HasChoice() {
		//end game
		if player.Choice == ROCK {
			if opponent.Choice == ROCK {
				player.SendResult("draw", opponent)
				opponent.SendResult("draw", player)

				return
			}
			if opponent.Choice == PAPER {
				player.SendResult("lose", opponent)
				opponent.SendResult("win", player)

				return
			}
			if opponent.Choice == SCISSORS {
				player.SendResult("win", opponent)
				opponent.SendResult("lose", player)

				return
			}
		}

		if player.Choice == PAPER {
			if opponent.Choice == ROCK {
				player.SendResult("win", opponent)
				opponent.SendResult("lose", player)

				return
			}
			if opponent.Choice == PAPER {
				player.SendResult("draw", opponent)
				opponent.SendResult("draw", player)

				return
			}
			if opponent.Choice == SCISSORS {
				player.SendResult("lose", opponent)
				opponent.SendResult("win", player)

				return
			}
		}

		if player.Choice == SCISSORS {
			if opponent.Choice == ROCK {
				player.SendResult("lose", opponent)
				opponent.SendResult("win", player)

				return
			}
			if opponent.Choice == PAPER {
				player.SendResult("win", opponent)
				opponent.SendResult("lose", player)

				return
			}
			if opponent.Choice == SCISSORS {
				player.SendResult("draw", opponent)
				opponent.SendResult("draw", player)

				return
			}
		}
	} else {
		data := map[string]string{
			"choice":  player.Choice.String(),
			"player1": player.Nick,
			"player2": "",
		}

		if opponent != nil {
			data["player2"] = opponent.Nick
		}

		errWS := player.SendMessage("templates/main.tmpl", "choice", data)
		if errWS != nil {
			log.Println("Error sending websocket response")
			return
		}

		return
	}
}

func connectPlayerToGame(game *Game, conn *websocket.Conn) error {
	if game.Player1.Conn == nil {
		if err := game.Player1.Connect(conn); err != nil {
			return errors.New(err.Error())
		}

		return nil
	} else {
		if err := game.Player2.Connect(conn); err != nil {
			return errors.New(err.Error())
		}

		// send message to players that another player has connected to the game
		errSendingMessage := game.Player1.SendMessage("templates/main.tmpl", "game", map[string]string{
			"opponent": game.Player2.Nick,
		})
		if errSendingMessage != nil {
			return errors.New(errSendingMessage.Error())
		}

		errSendingMessage2 := game.Player2.SendMessage("templates/main.tmpl", "game", map[string]string{
			"opponent": game.Player1.Nick,
		})
		if errSendingMessage2 != nil {
			return errors.New(errSendingMessage2.Error())
		}

		return nil
	}

}

func main() {
	logger := log.Default()
	r := gin.Default()
	r.Static("/images", "./images")
	r.Static("/components", "./components")
	r.Static("/styles", "./styles")
	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Rock Paper and Sissors Master",
			"rooms": Games,
		})
	})

	r.GET("/create-game-form", func(ctx *gin.Context) {
		html := template.Must(template.ParseFiles("templates/main.tmpl"))
		html.ExecuteTemplate(ctx.Writer, "gameform", nil)
	})

	r.GET("/join-game-form/:room", func(ctx *gin.Context) {
		room := ctx.Param("room")
		html := template.Must(template.ParseFiles("templates/main.tmpl"))
		html.ExecuteTemplate(ctx.Writer, "gameform", gin.H{"join": true, "room": room})
	})

	r.POST("/join-game/:room", func(ctx *gin.Context) {
		username := ctx.PostForm("nickname")
		var connected = createUser(username)

		currGame := ctx.Param("room")
		gameId, atoiErr := strconv.Atoi(currGame)
		if atoiErr != nil {
			log.Println("Cannot find game room!")
			return
		}

		var currentGame = findGame(gameId)
		currentGame.connectPlayer2(connected)
		logger.Printf("\nUser [%s] has joined to game %d", currentGame.Player2.Nick, currentGame.Id)

		html := template.Must(template.ParseFiles("templates/main.tmpl"))
		html.ExecuteTemplate(ctx.Writer, "waiting", gin.H{
			"room": currentGame.Id,
			"user": currentGame.Player2.Nick,
		})
	})

	r.GET("/start-game", func(ctx *gin.Context) {
		html := template.Must(template.ParseFiles("templates/main.tmpl"))
		html.ExecuteTemplate(ctx.Writer, "choose", gin.H{})
	})

	r.POST("/create-game", func(ctx *gin.Context) {
		username := ctx.PostForm("nickname")

		var connected = createUser(username)
		currGame := addNewGame(connected)
		logger.Printf("\nUser [%s] has created a new game %d", connected.Nick, currGame.Id)

		html := template.Must(template.ParseFiles("templates/main.tmpl"))
		html.ExecuteTemplate(ctx.Writer, "waiting", gin.H{
			"room": currGame.Id,
			"user": currGame.Player1.Nick,
		})
	})

	r.GET("/gameroom", func(ctx *gin.Context) {
		conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)

		if err != nil {
			return
		}

		defer conn.Close() //TODO research about defer

		currGame := ctx.Query("room")
		currUser := ctx.Query("user")
		gameId, atoiErr := strconv.Atoi(currGame)

		if atoiErr != nil {
			return
		}

		currentGame := findGame(gameId)

		if currentGame.isPlayerConnected(currUser) == false {
			connectionErr := connectPlayerToGame(currentGame, conn)
			if connectionErr != nil {
				logger.Fatal(connectionErr)
				return
			}

		}

		connectedUser := currentGame.connectedPlayer(currUser)
		logger.Printf(
			"\nUser [%s] has connected to game %d",
			connectedUser.Nick,
			currentGame.Id,
		)
		for {
			_, p, err := conn.ReadMessage()
			if err != nil {
				return
			}

			// get choice from the message
			message := WSMessage{}
			messageErr := json.Unmarshal(p, &message)
			if messageErr != nil {
				logger.Print("choice to struct error", err)
				return
			}

			if message.Choice == "" {
				logger.Print("Did not choose yet")
				return
			}

			// set player choice
			if connectedUser != nil {
				connectedUser.SetChoice(message.Choice)
				currentGame.handlePlayerMessage(connectedUser)
			}
		}
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}
