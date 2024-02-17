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

type User struct {
	Nick   string
	Choice string
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

func createUser(nickname string) *User {
	user := User{
		Nick:   nickname,
		Choice: "",
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

func (u *User) SetChoice(choice string) {
	u.Choice = choice
}

func (u *User) SendMessage(message string) error {
	err := u.Conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		return errors.New(err.Error())
	}
	return nil
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

		// send message to player 1 that another player has connected to his game
		var htmlTemplate bytes.Buffer

		if game.Player1.Choice != "" {
			tmpl := template.Must(template.ParseFiles("templates/main.tmpl"))
			if errExec := tmpl.ExecuteTemplate(&htmlTemplate, "choice", gin.H{
				"choice1": game.Player1.Choice,
				"player1": game.Player1.Nick,
				"player2": game.Player2.Nick,
			}); errExec != nil {
				return errors.New("Could not connect player to gamer!")
			}

			errSendingMessage := game.Player1.SendMessage(htmlTemplate.String())
			if errSendingMessage != nil {
				return errors.New(errSendingMessage.Error())
			}
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
		html.ExecuteTemplate(ctx.Writer, "choose", gin.H{
			"room":    currentGame.Id,
			"plyaer1": currentGame.Player2.Nick,
		})
	})

	r.POST("/create-game", func(ctx *gin.Context) {
		username := ctx.PostForm("nickname")

		var connected = createUser(username)
		currGame := addNewGame(connected)
		logger.Printf("\nUser [%s] has created a new game %d", connected.Nick, currGame.Id)

		html := template.Must(template.ParseFiles("templates/main.tmpl"))
		html.ExecuteTemplate(ctx.Writer, "choose", gin.H{
			"room":    currGame.Id,
			"player1": currGame.Player1.Nick,
			"player2": "",
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

			// set player choice
			if connectedUser != nil {
				connectedUser.SetChoice(message.Choice)
				opponent := currentGame.opponentPlayer(currUser)

				data := gin.H{
					"choice":  connectedUser.Choice,
					"player1": connectedUser.Nick,
					"player2": "",
				}

				if opponent != nil {
					data["player2"] = opponent.Nick
				}

				var htmlTemplate bytes.Buffer
				tmpl := template.Must(template.ParseFiles("templates/main.tmpl"))
				if errExec := tmpl.ExecuteTemplate(&htmlTemplate, "choice", data); errExec != nil {
					logger.Println("Error executing template", errExec)
					return
				}

				if errWS := connectedUser.SendMessage(htmlTemplate.String()); errWS != nil {
					logger.Println("Error sending websocket response")
					return
				}
			}
		}
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}
