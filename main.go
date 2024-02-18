package main

import (
	"encoding/json"
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
		gameId, atoiErr := strconv.Atoi(room)
		if atoiErr != nil {
			log.Println("Cannot find game room!")
			return
		}

		var currentGame = findGame(gameId)

		if currentGame.isGameEmpty(gameId) {
			html := template.Must(template.ParseFiles("templates/main.tmpl"))
			html.ExecuteTemplate(ctx.Writer, "gameform", gin.H{"join": true, "room": room})
			return
		} else {
			ctx.String(http.StatusForbidden, "Game is full")
			return
		}

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

		err := currentGame.connectPlayer2(connected)
		if err != nil {
			return
		}

		logger.Printf("\nUser [%s] has joined to game %d", currentGame.Player2.Nick, currentGame.Id)

		html := template.Must(template.ParseFiles("templates/main.tmpl"))
		html.ExecuteTemplate(ctx.Writer, "waiting", gin.H{
			"room": currentGame.Id,
			"user": currentGame.Player2.Nick,
		})
	})

	r.GET("/play-again", func(ctx *gin.Context) {
	})

	r.GET("/quit-game", func(ctx *gin.Context) {
		currGame := ctx.Query("room")
		currUser := ctx.Query("user")
		gameId, atoiErr := strconv.Atoi(currGame)

		if atoiErr != nil {
			ctx.String(http.StatusNotFound, "Error trying to find current game")
		}

		currentGame := findGame(gameId)
		err := currentGame.disconnectPlayer(currUser)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Error trying to disconnect from game")
		}

		html := template.Must(template.ParseFiles("templates/main.tmpl"))
		html.ExecuteTemplate(ctx.Writer, "main", gin.H{
			"rooms": Games,
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
			connectionErr := currentGame.connectPlayerToGame(conn)
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
