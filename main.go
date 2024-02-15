package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var score = 0

type User struct {
	Nick  string
	Score int
}

type Game struct {
	Id       int
	Creator  string
	Opponent string
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
	r := gin.Default()
	r.Static("/images", "./images")
	r.Static("/components", "./components")
	r.Static("/styles", "./styles")
	r.LoadHTMLGlob("templates/*")

	var ConnectedUsers = []User{{Nick: "dummy", Score: 0}}
	var Games = []Game{{Id: 0, Creator: "dummy"}}

	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Rock Paper and Sissors Master",
			"score": score,
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
		nickname := ctx.PostForm("nickname")

		var connected User
		if len(ConnectedUsers) == 0 {
			connected = User{Nick: nickname, Score: 0}
			ConnectedUsers = append(ConnectedUsers, connected)
		} else {
			for _, con := range ConnectedUsers {
				if con.Nick == nickname {
					connected = con
					break
				}
			}

			if (connected == User{}) {
				connected = User{Nick: nickname, Score: 0}
				ConnectedUsers = append(ConnectedUsers, connected)
			}
		}

		currGame := ctx.Param("room")

		gameId, atoiErr := strconv.Atoi(currGame)
		if atoiErr != nil {
			log.Println("Cannot find game room!")
			return
		}

		html := template.Must(template.ParseFiles("templates/main.tmpl"))
		html.ExecuteTemplate(ctx.Writer, "choose", gin.H{
			"room": gameId,
			"user": nickname,
		})
	})

	r.POST("/create-game", func(ctx *gin.Context) {
		nickname := ctx.PostForm("nickname")
		var connected User
		if len(ConnectedUsers) == 0 {
			connected = User{Nick: nickname, Score: 0}
			ConnectedUsers = append(ConnectedUsers, connected)
		} else {
			for _, con := range ConnectedUsers {
				if con.Nick == nickname {
					connected = con
					break
				}
			}

			if (connected == User{}) {
				connected = User{Nick: nickname, Score: 0}
				ConnectedUsers = append(ConnectedUsers, connected)
			}
		}

		currGame := Game{Id: len(Games), Creator: connected.Nick}
		Games = append(Games, currGame)

		html := template.Must(template.ParseFiles("templates/main.tmpl"))
		html.ExecuteTemplate(ctx.Writer, "choose", gin.H{
			"room": currGame.Id,
			"user": nickname,
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
		// new client joined to room gameId
		fmt.Printf("User %s joined room id %d", currUser, gameId)
		for {
			msgType, p, err := conn.ReadMessage()
			if err != nil {
				return
			}
			fmt.Printf("\nReceived message: {%s}", string(p))

			// get choice from the message
			choice := WSMessage{}
			choiceErr := json.Unmarshal(p, &choice)

			if choiceErr != nil {
				fmt.Print("choice to struct error", err)
				return
			}

			fmt.Printf("\nUser choice %s", choice.Choice)

			if err := conn.WriteMessage(msgType, p); err != nil {
				return
			}
		}
	})

	r.GET("/choose/:choice", func(ctx *gin.Context) {
		choice := ctx.Param("choice")

		log.Println("USER picked: ", choice)
		ctx.HTML(http.StatusOK, "game.tmpl", gin.H{
			"choice": choice,
		})
	})

	r.GET("/result", func(ctx *gin.Context) {
		userChoice := ctx.Query("user")
		choices := [3]string{"rock", "paper", "scissors"}
		random := rand.Intn(len(choices))

		if choices[random] == userChoice {
			random = rand.Intn(len(choices))
		}

		if choices[random] == "rock" {
			if userChoice == "paper" {
				log.Println("User win!")
				score = score + 1
				ctx.HTML(http.StatusOK, "win.tmpl", gin.H{
					"choice": userChoice,
					"house":  choices[random],
					"score":  score,
				})
				return
			} else {
				log.Println("User lose!")
				if score > 0 {
					score = score - 1
				}
				ctx.HTML(http.StatusOK, "lose.tmpl", gin.H{
					"choice": userChoice,
					"house":  choices[random],
					"score":  score,
				})
				return
			}
		}

		if choices[random] == "paper" {
			if userChoice == "scissors" {
				log.Println("User win!")
				score = score + 1
				ctx.HTML(http.StatusOK, "win.tmpl", gin.H{
					"choice": userChoice,
					"house":  choices[random],
					"score":  score,
				})
				return
			} else {
				log.Println("User lose!")
				if score > 0 {
					score = score - 1
				}
				ctx.HTML(http.StatusOK, "lose.tmpl", gin.H{
					"choice": userChoice,
					"house":  choices[random],
					"score":  score,
				})
				return
			}
		}

		if choices[random] == "scissors" {
			if userChoice == "rock" {
				log.Println("User win!")
				score = score + 1
				ctx.HTML(http.StatusOK, "win.tmpl", gin.H{
					"choice": userChoice,
					"house":  choices[random],
					"score":  score,
				})
				return
			} else {
				log.Println("User lose!")
				if score > 0 {
					score = score - 1
				}
				ctx.HTML(http.StatusOK, "lose.tmpl", gin.H{
					"choice": userChoice,
					"house":  choices[random],
					"score":  score,
				})
				return
			}
		}

	})

	r.GET("/main/:room", func(ctx *gin.Context) {
		html := template.Must(template.ParseFiles("templates/index.tmpl"))
		html.ExecuteTemplate(ctx.Writer, "main", Game{})
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}
