package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"

	"github.com/gorilla/websocket"
)

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

	if choice == "none" {
		u.Choice = 0
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

func (u *User) SendResult(result string, opponent *User, game int) {
	gameId := fmt.Sprint(game)

	if result == "draw" {
		if errPlayer := u.SendTemplate("templates/draw.tmpl", map[string]string{
			"choice":   u.Choice.String(),
			"opponent": opponent.Nick,
			"house":    opponent.Choice.String(),
			"room":     gameId,
			"user":     u.Nick,
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
			"room":     gameId,
			"user":     u.Nick,
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
			"room":     gameId,
			"user":     u.Nick,
		}); errPlayer != nil {
			return
		}

		return
	}

}

func (g *Game) connectPlayer2(player *User) error {
	if player.Conn != nil {
		return errors.New("Player already connected")
	}

	g.Player2 = player
	return nil
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

func (g Game) isGameEmpty(id int) bool {
	if g.Player2 != nil {
		return false
	}
	return true
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
				player.SendResult("draw", opponent, g.Id)
				opponent.SendResult("draw", player, g.Id)

				return
			}
			if opponent.Choice == PAPER {
				player.SendResult("lose", opponent, g.Id)
				opponent.SendResult("win", player, g.Id)

				return
			}
			if opponent.Choice == SCISSORS {
				player.SendResult("win", opponent, g.Id)
				opponent.SendResult("lose", player, g.Id)

				return
			}
		}

		if player.Choice == PAPER {
			if opponent.Choice == ROCK {
				player.SendResult("win", opponent, g.Id)
				opponent.SendResult("lose", player, g.Id)

				return
			}
			if opponent.Choice == PAPER {
				player.SendResult("draw", opponent, g.Id)
				opponent.SendResult("draw", player, g.Id)

				return
			}
			if opponent.Choice == SCISSORS {
				player.SendResult("lose", opponent, g.Id)
				opponent.SendResult("win", player, g.Id)

				return
			}
		}

		if player.Choice == SCISSORS {
			if opponent.Choice == ROCK {
				player.SendResult("lose", opponent, g.Id)
				opponent.SendResult("win", player, g.Id)

				return
			}
			if opponent.Choice == PAPER {
				player.SendResult("win", opponent, g.Id)
				opponent.SendResult("lose", player, g.Id)

				return
			}
			if opponent.Choice == SCISSORS {
				player.SendResult("draw", opponent, g.Id)
				opponent.SendResult("draw", player, g.Id)

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

func (game *Game) connectPlayerToGame(conn *websocket.Conn) error {
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

func (game *Game) disconnectPlayer(username string) error {
	if username == game.Player1.Nick {
		game.Player1 = game.Player2
		game.Player2 = nil

		err1 := game.Player1.SendMessage("templates/main.tmpl", "wait", map[string]string{})
		if err1 != nil {
			return err1
		}

		return nil
	}

	if username == game.Player2.Nick {
		game.Player2 = nil
		if game.Player1 != nil {
			err := game.Player1.SendMessage("templates/main.tmpl", "wait", map[string]string{})
			if err != nil {
				return err
			}

		}
		return nil
	}

	return errors.New("Could not find player in game")
}
