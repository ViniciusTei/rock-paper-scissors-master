package main

import (
	"log"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
)

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

	r.GET("/choose/:choice", func(ctx *gin.Context) {
		choice := ctx.Param("choice")

		log.Println("USER picked: ", choice)
		ctx.HTML(http.StatusOK, "game.tmpl", gin.H{
			"choice": choice,
		})
	})

	r.GET("/result", func(ctx *gin.Context) {
		choices := [3]string{"rock", "paper", "scissors"}
		random := rand.Intn(len(choices))
		log.Println("Rando choice", choices[random])
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}
