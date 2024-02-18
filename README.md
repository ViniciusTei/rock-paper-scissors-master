# Frontend Mentor - Rock, Paper, Scissors solution

This is a solution to the [Rock, Paper, Scissors challenge on Frontend Mentor](https://www.frontendmentor.io/challenges/rock-paper-scissors-game-pTgwgvgH). But in this challenge I tried to focus on back-end with Go lang. Making a multiplayer online game of Rock, Paer, Scissors.

![Preview](/design/desktop-preview.jpg)

## Table of contents

- [Overview](#overview)
  - [The challenge](#the-challenge)
  - [Links](#links)
- [My process](#my-process)
  - [Built with](#built-with)
  - [What I learned](#what-i-learned)
  - [Continued development](#continued-development)
- [Author](#author)

## Overview

### The challenge

Users should be able to:

- Create a game of Rock, Paper, Scissors
- Join an existing game 
- Play Rock, Paper, Scissors against the user
- Maintain the state of the score after refreshing the browser (TODO)

### Links

(TODO)
- Solution URL: [Add solution URL here](https://your-solution-url.com)
- Live Site URL: [Add live site URL here](https://your-live-site-url.com)

## My process

In this challenge I created a game of Rock, Paper and Scissors that the players can crate a room to play against other players. The player that creates the room waits the second player to join so they can choose their options, when both
players have choosen their options the game show them the result.

### Built with

- [HTMX](https://htmx.org/) - For client and server comunication
- [Go](https://go.dev/) - Main development language
- [Gin](https://gin-gonic.com/docs/) - For routing and server development
- [Websocket](https://github.com/gorilla/websocket) - For real time comunication

### What I learned

- HTMX syntax to comunicate with the server using basic HTTP requests and WebSocket comunication;
- Basic Go Lang syntax, like Types, Structs, Functions, Pointers, etc;

### Continued development

- [ ] Save to a Sqlite database user and game data
- [ ] Deploy to the internet

## Author

<table>
  <tr>
    <td align="center">
      <a href="#">
        <img src="https://github.com/ViniciusTei.png" width="100px;" alt="Foto do ViniciusTei"/><br>
        <sub>
          <b>ViniciusTei</b>
        </sub>
      </a>
    </td>
</table>


