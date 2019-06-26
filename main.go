package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/option"
)

var gcp *datastore.Client
var gcpErr error
var ctx = context.Background()
var quit = make(chan struct{})
var wumpus []Wumpus
var keys []*datastore.Key
var err error

//Wumpus boi
type Wumpus struct {
	Credits   int
	Name      string
	Color     int
	Age       int
	Health    int
	Hunger    int
	Energy    int
	Happiness int
	Sick      bool
	Sleeping  bool
	Left      bool
}

func main() {
	gcp, gcpErr = datastore.NewClient(ctx, "wumpagotchi", option.WithCredentialsFile("./WumpagotchiCredentials.json"))
	if gcpErr != nil {
		fmt.Println("Time Wizard Wump failed to wake up")
		os.Exit(1)
	}

	go timespell()
	go agespell()

	fmt.Println("Time Wizard Wump has awoken!\nRunning until slain by a termination singal ...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func timespell() {
	for range time.Tick(time.Hour * 2) {
		updateList()
		fmt.Println("Updated List")
		for w := 0; w+1 <= len(wumpus); w++ {
			rand.Seed(time.Now().UnixNano())
			//First Check if the Wumpus left so we don't run anything else if they have
			if wumpus[w].Left == true {
				continue
			}
			if wumpus[w].Energy > 8 && wumpus[w].Happiness > 8 && wumpus[w].Health > 8 && wumpus[w].Hunger > 8 && wumpus[w].Sick == false && wumpus[w].Age > 1 {
				wumpus[w].Credits += 10
			}
			//Have a 10% chance of making the Wumpus sick
			if rand.Float32() < 0.10 {
				wumpus[w].Sick = true
			}
			//Next check if they are sick and if they are reduce all stats by 1
			if wumpus[w].Sick == true {
				wumpus[w].Health--
				wumpus[w].Energy--
				wumpus[w].Hunger--
				wumpus[w].Happiness--
			}
			//Next have a 25% chance of reducing happiness
			if rand.Float32() < 0.25 {
				wumpus[w].Happiness--
			}
			//Have a 50% chance of reducing hunger by 1
			if rand.Float32() < 0.50 {
				wumpus[w].Hunger--
			}
			//Check if they are sleeping if they are add 4 to energy and 1 to health
			//then if have a 75% chance of reducing energy by 1
			//If they are not sleeping and their energy is at or below 0 mark them as sleeping
			if wumpus[w].Sleeping == true {
				wumpus[w].Energy += 4
				wumpus[w].Health++
			} else if rand.Float32() < 0.75 {
				wumpus[w].Energy--
			} else if wumpus[w].Energy <= 0 {
				wumpus[w].Sleeping = true
			}
			//Check if hunger is @ or below 0 if so reduce health and happiness by 1
			if wumpus[w].Hunger <= 0 {
				wumpus[w].Health--
				wumpus[w].Happiness--
			}

			//Make sure that all values are set to 0 before checking for happiness and writing to GCP Datastore
			if wumpus[w].Health < 0 {
				wumpus[w].Health = 0
			}
			if wumpus[w].Hunger < 0 {
				wumpus[w].Hunger = 0
			}
			if wumpus[w].Energy < 0 {
				wumpus[w].Energy = 0
			}
			if wumpus[w].Happiness < 0 {
				wumpus[w].Happiness = 0
			}
			if wumpus[w].Health > 10 {
				wumpus[w].Health = 10
			}
			if wumpus[w].Hunger > 10 {
				wumpus[w].Hunger = 10
			}
			if wumpus[w].Energy > 10 {
				wumpus[w].Energy = 10
			}
			if wumpus[w].Happiness > 10 {
				wumpus[w].Happiness = 10
			}
			//Check if the happiness is @ or below 0 and then have a 50% chance if both are true mark the wumpus as Left
			if wumpus[w].Happiness <= 0 && rand.Float32() < 0.50 {
				wumpus[w].Left = true
			}

			userKey := datastore.NameKey("User", keys[w].Name, nil)
			if _, err := gcp.Put(ctx, userKey, &wumpus[w]); err != nil {
				fmt.Println("==Warning==\nFailed to add Wumpus to Datastore")
				break
			}
		}
	}
}

func agespell() {
	for range time.Tick(time.Hour * 24) {
		updateList()
		fmt.Println("Updated List")
		for w := 0; w+1 <= len(wumpus); w++ {
			if wumpus[w].Left == true {
				continue
			}
			if wumpus[w].Age == 14 {
				wumpus[w].Left = true
				continue
			}
			wumpus[w].Age++
			userKey := datastore.NameKey("User", keys[w].Name, nil)
			if _, err := gcp.Put(ctx, userKey, &wumpus[w]); err != nil {
				fmt.Println("==Warning==\nFailed to add Wumpus to Datastore")
				break
			}
		}
	}
}

func updateList() {
	query := datastore.NewQuery("User")
	wumpus = nil
	keys, err = gcp.GetAll(ctx, query, &wumpus)
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err.Error())
	}
}
