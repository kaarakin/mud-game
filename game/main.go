package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Player представляет игрока в игре.
type Player struct {
	container PlayerContainer
	location  *Room
}

// Description описывает характеристики комнаты.
type Description struct {
	goCase      string
	lookCase    string
	area        string
	prefix      string
	goalPrefix  string
	locker      string
	keyType     string
	emptyMsg    string
	lockedMsg   string
	unlockedMsg string
}

// Room представляет комнату в игре.
type Room struct {
	name        string
	description Description
	containers  []RoomContainer
	passages    []*Room
	isLocked    bool
	goals       []Goal
}

// Item представляет предмет в игре.
type Item struct {
	name, itemType string
}

// PlayerContainer представляет инвентарь игрока.
type PlayerContainer struct {
	isPuttedOn bool
	inventory  []Item
}

// RoomContainer представляет контейнер (инвентарь) в комнате.
type RoomContainer struct {
	prefix    string
	inventory []Item
}

// Goal представляет цель в игре.
type Goal struct {
	title      string
	isAchieved bool
	сheck      func(player Player, goalRoom *Room)
}

// lookAround возвращает описание текущей комнаты и доступные действия.
func lookAround(player Player, isGoCase bool) string {
	fullAnswer := make([]string, 0)
	answer := make([]string, 0)
	emptyContainers := true

	// Проверка наличия описания для осмотра и движения.
	// [lookCase] / [goCase]
	if lookCase := player.location.description.lookCase; lookCase != "" && !isGoCase {
		answer = append(answer, lookCase)
	} else if goCase := player.location.description.goCase; goCase != "" && isGoCase {
		answer = append(answer, goCase)
	}

	// Описание инвентаря в комнате.
	if !isGoCase {
		// [inventoryMsg]
		if len(player.location.containers) > 0 {
			inventoryMsg := make([]string, 0)

			for _, container := range player.location.containers {
				inventoryList := make([]string, 0)
				items := make([]string, 0)

				for _, item := range container.inventory {
					// item -> [items]
					items = append(items, item.name)
				}

				if len(items) > 0 {
					emptyContainers = false

					// [prefix] -> [inventoryList]
					inventoryList = append(inventoryList, container.prefix)

					// [items] -> [inventoryList]
					inventoryList = append(inventoryList, strings.Join(items, ", "))

					// [inventoryList] -> [inventoryMsg]
					inventoryMsg = append(inventoryMsg, strings.Join(inventoryList, ": "))
				}
			}

			if emptyContainers {
				answer = append(answer, player.location.description.emptyMsg)
			} else {
				answer = append(answer, strings.Join(inventoryMsg, ", "))
			}
		}

		// Описание целей в комнате.
		// [goal]
		if len(player.location.goals) > 0 {
			goalsMsg := make([]string, 0)
			goals := make([]string, 0)

			for _, goal := range player.location.goals {
				if !goal.isAchieved {
					goals = append(goals, goal.title)
				}
			}

			// [goalPrefix]
			goalsMsg = append(goalsMsg, player.location.description.goalPrefix)
			goalsMsg = append(goalsMsg, strings.Join(goals, " и "))

			answer = append(answer, strings.Join(goalsMsg, " "))
		}
	}

	fullAnswer = append(fullAnswer, strings.Join(answer, ", "))

	// Описание доступных выходов из комнаты.
	if len(player.location.passages) > 0 {
		passagesMsg := make([]string, 0)
		passagesMsg = append(passagesMsg, "можно пройти")
		availablePassages := make([]string, 0)

		for _, passage := range player.location.passages {
			if passage.description.area == player.location.description.area {
				availablePassages = append(availablePassages, passage.name)
			} else {
				availablePassages = append(availablePassages, passage.description.area)
			}

		}

		passagesMsg = append(passagesMsg, strings.Join(availablePassages, ", "))
		fullAnswer = append(fullAnswer, strings.Join(passagesMsg, " - "))
	}
	return strings.Join(fullAnswer, ". ")
}

// checkGoals проверяет достижение целей в комнате.
func checkGoals(room *Room, goals ...*Goal) {
	for _, goal := range goals {
		goal.сheck(player, room)
	}
}

// handleLookCommand обрабатывает команду "осмотреться".
func handleLookCommand() string {
	return lookAround(player, false)
}

// handleGoCommand обрабатывает команду "идти".
func handleGoCommand(parsedCommand []string) string {
	if len(parsedCommand) != 2 {
		return "неверное количество аргументов для команды 'идти'"
	}

	destination := parsedCommand[1]

	for _, passage := range player.location.passages {
		if passage.isLocked {
			return passage.description.lockedMsg
		}

		if destination == passage.name || destination == passage.description.area {
			player.location = passage
			return lookAround(player, true)
		}
	}

	return "нет пути в " + destination
}

// handleWearCommand обрабатывает команду "надеть".
func handleWearCommand(parsedCommand []string) string {
	if len(parsedCommand) != 2 {
		return "неверное количество аргументов для команды 'надеть'"
	}

	wearedItem := parsedCommand[1]

	for idxContainer, container := range player.location.containers {
		for idxItem, item := range container.inventory {
			if item.itemType == "playerContainer" && item.name == wearedItem {
				player.container.isPuttedOn = true
				player.location.containers[idxContainer].inventory = append(player.location.containers[idxContainer].inventory[:idxItem], player.location.containers[idxContainer].inventory[idxItem+1:]...)
				return "вы надели: " + item.name
			}
		}
	}

	return "нечего надеть"
}

// handleTakeCommand обрабатывает команду "взять".
func handleTakeCommand(parsedCommand []string) string {
	if len(parsedCommand) != 2 {
		return "неверное количество аргументов для команды 'взять'"
	}

	if !player.container.isPuttedOn {
		return "некуда класть"
	}

	takenItem := parsedCommand[1]

	for idxContainer, container := range player.location.containers {
		for idxItem, item := range container.inventory {
			if (item.itemType == "item" || item.itemType == "key") && item.name == takenItem {
				player.container.inventory = append(player.container.inventory, item)
				player.location.containers[idxContainer].inventory = append(player.location.containers[idxContainer].inventory[:idxItem], player.location.containers[idxContainer].inventory[idxItem+1:]...)
				return "предмет добавлен в инвентарь: " + item.name
			}
		}
	}

	return "нет такого"
}

// handleApplyCommand обрабатывает команду "применить".
func handleApplyCommand(parsedCommand []string) string {
	if len(parsedCommand) != 3 {
		return "неверное количество аргументов для команды 'применить'"
	}

	appliedItem := parsedCommand[1]

	for _, item := range player.container.inventory {
		if item.name == appliedItem {
			for _, location := range player.location.passages {
				if (location.description.keyType == item.itemType) && location.isLocked {
					location.isLocked = false
					return location.description.unlockedMsg
				}
			}
			return "не к чему применить"
		}
	}

	return "нет предмета в инвентаре - " + appliedItem
}

var (
	player Player
	world  map[string]*Room
)

func main() {
	initGame()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Игра начата")
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("Было введено: %q\n", line)
		if line == "завершить" {
			fmt.Println("Завершение игры...")
			break
		}
		answer := handleCommand(line)
		fmt.Println(answer)
	}
}

// initGame инициализирует игровой мир и игрока.
func initGame() {
	var kitchen Room
	var corridor Room
	var room Room
	var street Room

	corridor = Room{
		name: "коридор",
		description: Description{
			goCase: "ничего интересного",
			area:   "домой",
		},
		containers: []RoomContainer{},
		passages: []*Room{
			&kitchen, &room, &street,
		},
		isLocked: false,
		goals:    []Goal{},
	}

	room = Room{
		name: "комната",
		description: Description{
			goCase:   "ты в своей комнате",
			area:     "домой",
			emptyMsg: "пустая комната",
		},
		containers: []RoomContainer{
			{
				"на столе",
				[]Item{
					{
						"ключи", "key",
					},
					{
						"конспекты", "item",
					},
				},
			},
			{
				"на стуле",
				[]Item{
					{
						"рюкзак", "playerContainer",
					},
				},
			},
		},
		passages: []*Room{
			&corridor,
		},
		isLocked: false,
		goals:    []Goal{},
	}

	street = Room{
		name: "улица",
		description: Description{
			goCase:      "на улице весна",
			locker:      "дверь",
			keyType:     "key",
			lockedMsg:   "дверь закрыта",
			unlockedMsg: "дверь открыта",
			area:        "улица",
		},
		containers: []RoomContainer{},
		passages: []*Room{
			&corridor,
		},
		isLocked: true,
		goals:    []Goal{},
	}

	kitchen = Room{
		name: "кухня",
		description: Description{
			lookCase:   "ты находишься на кухне",
			goCase:     "кухня, ничего интересного",
			prefix:     "на столе",
			goalPrefix: "надо",
			area:       "домой",
		},
		containers: []RoomContainer{
			{
				"на столе",
				[]Item{
					{
						"чай",
						"item",
					},
				},
			},
		},
		passages: []*Room{
			&corridor,
		},
		isLocked: false,
		goals: []Goal{
			{
				"собрать рюкзак",
				false,
				func(player Player, kitchen *Room) {
					keysTaken := false
					notesTaken := false
					for _, item := range player.container.inventory {
						if item.name == "ключи" {
							keysTaken = true
						}
						if item.name == "конспекты" {
							notesTaken = true
						}
						if keysTaken && notesTaken {
							kitchen.goals[0].isAchieved = true
						}
					}
				},
			},
			{
				"идти в универ",
				false,
				func(player Player, kitchen *Room) {
					if kitchen.goals[0].isAchieved && player.location.name == "улица" {
						kitchen.goals[1].isAchieved = true
					}
				},
			},
		},
	}

	world = map[string]*Room{
		"кухня":   &kitchen,
		"коридор": &corridor,
		"комната": &room,
		"улица":   &street,
	}

	player = Player{
		container: PlayerContainer{
			false,
			[]Item{},
		},
		location: &kitchen,
	}
}

// handleCommand обрабатывает введенную команду.
func handleCommand(command string) string {
	checkGoals(world["кухня"], &world["кухня"].goals[0], &world["кухня"].goals[1])

	parsedCommand := strings.Split(command, " ")
	action := parsedCommand[0]

	switch action {
	case "осмотреться":
		return handleLookCommand()

	case "идти":
		return handleGoCommand(parsedCommand)

	case "надеть":
		return handleWearCommand(parsedCommand)

	case "взять":
		return handleTakeCommand(parsedCommand)

	case "применить":
		return handleApplyCommand(parsedCommand)
	}

	return "неизвестная команда"
}

