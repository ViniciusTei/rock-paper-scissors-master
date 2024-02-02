package main

import (
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var score = 0

func main() {
	r := gin.Default()
	r.Static("/images", "./images")
	r.Static("/components", "./components")
	r.Static("/styles", "./styles")
	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Rock Paper and Sissors Master",
		})
	})

	r.GET("/score", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, strconv.FormatInt(int64(score), 10))
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
				ctx.HTML(http.StatusOK, "win.tmpl", gin.H{
					"choice": userChoice,
					"house":  choices[random],
				})
				score = score + 1
				return
			} else {
				log.Println("User lose!")
				ctx.HTML(http.StatusOK, "lose.tmpl", gin.H{
					"choice": userChoice,
					"house":  choices[random],
				})

				if score > 0 {
					score = score - 1
				}
				return
			}
		}

		if choices[random] == "paper" {
			if userChoice == "scissors" {
				log.Println("User win!")
				ctx.HTML(http.StatusOK, "win.tmpl", gin.H{
					"choice": userChoice,
					"house":  choices[random],
				})
				score = score + 1
				return
			} else {
				log.Println("User lose!")
				ctx.HTML(http.StatusOK, "lose.tmpl", gin.H{
					"choice": userChoice,
					"house":  choices[random],
				})
				if score > 0 {
					score = score - 1
				}
				return
			}
		}

		if choices[random] == "scissors" {
			if userChoice == "rock" {
				log.Println("User win!")
				ctx.HTML(http.StatusOK, "win.tmpl", gin.H{
					"choice": userChoice,
					"house":  choices[random],
				})
				score = score + 1
				return
			} else {
				log.Println("User lose!")
				ctx.HTML(http.StatusOK, "lose.tmpl", gin.H{
					"choice": userChoice,
					"house":  choices[random],
				})
				if score > 0 {
					score = score - 1
				}
				return
			}
		}

	})

	r.GET("/main", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "main.tmpl", gin.H{})
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}
