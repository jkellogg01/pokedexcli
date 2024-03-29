package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/jkellogg01/pokedexcli/internal/pokeapi"
)

const prompt string = "\npokedex > "

type cliCommand struct {
	name string
	desc string
	hand func(*config, []string) error
}

type config struct {
	helpMsg string
	nextLoc string
	prevLoc string
	api     pokeapi.ApiService
}

func main() {
	verbose := flag.Bool("v", false, "verbose output will display debug logging")
	flag.Parse()
	if *verbose {
		log.SetLevel(log.DebugLevel)
	}
	log.Debug("Startup sequence begin:")
	commands := map[string]cliCommand{
		"help": {
			name: "help",
			desc: "displays a help message",
			hand: handleHelp,
		},
		"exit": {
			name: "exit",
			desc: "exits the pokedex REPL",
			hand: handleExit,
		},
		"map": {
			name: "map",
			desc: "advances the map, displaying location information",
			hand: handleMap,
		},
		"mapb": {
			name: "mapb",
			desc: "moves the map backwards, displaying location information",
			hand: handleMapb,
		},
		"explore": {
			name: "explore",
			desc: "explores a location-area, revealing nearby pokemon",
			hand: handleExplore,
		},
		"catch": {
			name: "catch",
			desc: "attempts to catch a pokemon",
			hand: handleCatch,
		},
	}
	log.Debug("Command map created")
	cfg := initConfig(commands)
	log.Debug("Config struct initialized")
	log.Print("Starting pokedex cli...\nType 'help' for more information")
	input := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(prompt)
		if !input.Scan() {
			err := input.Err()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Please type a command.\nType 'help' for more information")
		}
		cmd, args := readCommand(input.Text())
		if c, ok := commands[cmd]; ok {
			log.Debug("", "command", cmd, "args", args)
			err := c.hand(cfg, args)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			fmt.Println("Invalid command.\nType 'help' for more information")
		}
	}
}

func initConfig(commands map[string]cliCommand) *config {
	// TODO: maps are unordered, there should
	// be a better way of organizing this.
	var helpMsg string
	for k, v := range commands {
		log.Debugf("Registering command: %s", k)
		helpMsg += fmt.Sprintf("%8s -- %s\n", k, v.desc)
	}
	log.Debug("Initializing api service")
	api := pokeapi.NewApiService(time.Minute * 3)
	return &config{
		nextLoc: "",
		prevLoc: "",
		helpMsg: helpMsg,
		api:     api,
	}
}

func readCommand(in string) (string, []string) {
	vals := strings.Split(in, " ")
	if len(vals) <= 1 {
		return in, []string{}
	}
	return vals[0], vals[1:]
}

func handleMap(cfg *config, args []string) error {
	url := cfg.nextLoc
	data, err := cfg.api.GetLocations(url)
	if err != nil {
		return err
	}
	if data.Next != nil {
		cfg.nextLoc = *data.Next
	} else {
		cfg.nextLoc = ""
	}
	if data.Prev != nil {
		cfg.prevLoc = *data.Prev
	} else {
		cfg.prevLoc = ""
	}
	for _, location := range data.Results {
		fmt.Println(location.Name)
	}
	return nil
}

func handleMapb(cfg *config, args []string) error {
	url := cfg.prevLoc
	if url == "" {
		fmt.Println("Error: already on first page.")
		return nil
	}
	data, err := cfg.api.GetLocations(url)
	if err != nil {
		return err
	}
	if data.Next != nil {
		cfg.nextLoc = *data.Next
	} else {
		cfg.nextLoc = ""
	}
	if data.Prev != nil {
		cfg.prevLoc = *data.Prev
	} else {
		cfg.prevLoc = ""
	}
	for _, location := range data.Results {
		fmt.Println(location.Name)
	}
	return nil
}

func handleExplore(cfg *config, args []string) error {
	if len(args) != 1 {
		fmt.Println("Usage: explore <location_area>\nType 'help' for more information")
		return nil
	}
	fmt.Printf("Exploring %s...\n", args[0])
	pkmns, err := cfg.api.GetLocationPkmn(args[0])
	if err != nil {
		return err
	}
	fmt.Println("Found Pokemon:")
	for _, pkmn := range pkmns {
		fmt.Printf("- %s\n", pkmn)
	}
	return nil
}

func handleCatch(cfg *config, args []string) error {
	if len(args) != 1 {
		fmt.Println("Usage: catch <pokemon>\nType 'help' for more information")
		return nil
	}
	fmt.Printf("Throwing a Pokeball at %s...\n", args[0])
	pkmn, err := cfg.api.GetPkmn(args[0])
	if err != nil {
		return err
	}
	catch := rand.Intn(pkmn.BaseXP)
	if int(math.Pow(float64(catch), 2)) > pkmn.BaseXP {
		// catch logic: add to pokedex
		fmt.Printf("%s was caught!", pkmn.Name)
	} else {
		fmt.Printf("%s escaped!", pkmn.Name)
	}
	return nil
}

func handleHelp(cfg *config, args []string) error {
	if cfg.helpMsg == "" {
		return errors.New("help message unexpectedly empty")
	}
	fmt.Println(cfg.helpMsg)
	return nil
}

func handleExit(cfg *config, args []string) error {
	os.Exit(0)
	return fmt.Errorf("os.Exit did not work...")
}
