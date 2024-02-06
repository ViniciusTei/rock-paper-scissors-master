package main

import (
	"log"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
)

var score = 0

type User struct {
	Nick  string
	Score int
}

type Game struct {
	Creator  string
	Opponent string
}

func main() {
	r := gin.Default()
	r.Static("/images", "./images")
	r.Static("/components", "./components")
	r.Static("/styles", "./styles")
	r.LoadHTMLGlob("templates/*")

	var ConnectedUsers []User
	var Games []Game

	r.GET("/", func(ctx *gin.Context) {

		ctx.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Rock Paper and Sissors Master",
			"score": score,
		})
	})

	r.GET("/create-game-form", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "create-game-form.tmpl", gin.H{})
	})

	r.POST("/create-game", func(ctx *gin.Context) {
		nickname := ctx.PostForm("nickname")
		connected := User{Nick: nickname, Score: 0}
		ConnectedUsers = append(ConnectedUsers, connected)
		Games = append(Games, Game{Creator: connected.Nick})

		ctx.HTML(http.StatusOK, "main.tmpl", gin.H{
			"score": connected.Score,
			"user":  connected.Nick,
		})
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

		for choices[random] == userChoice {
			log.Println("Draw")
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

	r.GET("/main", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "main.tmpl", gin.H{
			"score": score,
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}
