package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/cryptidcodes/gator/internal/config"
	"github.com/cryptidcodes/gator/internal/database"

	_ "github.com/lib/pq"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

func main() {
	// read the config file
	cfg, err := config.Read()
	if err != nil {
		println(err)
	}

	// open the database connection
	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatal("failed to open database")
	}
	if err = db.Ping(); err != nil {
		log.Fatal("failed to ping database")
	}
	
	dbQueries := database.New(db)
	if dbQueries == nil {
		log.Fatal("failed to create database queries")
	}
	
	// create a new state
	s := state{
		db:  dbQueries,
		cfg: &cfg,
	}
	
	// create an instance of the commands struct and initialize the cmdmap
	cmds := commands{
		cmdmap: make(map[string]func(*state, command) error),
	}

	// register commands
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerGetUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))
	
	// confirm the user input at least two args. Example: gator login
	if len(os.Args) < 2 {
		log.Fatal("you must input a command to use gator\n")
	}

	// confirm the command the user is trying to run exists
	_, exists := cmds.cmdmap[os.Args[1]]
	if !exists {
		log.Fatal("unregistered command, please use a registered command\n")
	}

	// build the func
	userCmd := command{
		Name: os.Args[1],
		Args: make([]string, 0),
	}

	// add args
	if len(os.Args) > 2 {
		userCmd.Args = os.Args[2:]
	}

	// run the func
	err = cmds.run(&s, userCmd)
	if err != nil {
		log.Fatal(err)
	}
}