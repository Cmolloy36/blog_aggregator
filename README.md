# RSS Blog Aggregator

## Purpose
A CLI tool to obtain information about RSS feeds.

## Required installations:
* Go
* Postgres

## Setting up config file
Create the file `~/.gatorconfig.json`. In it, create a JSON object with the following fields:
```
{
    "db_url": "postgres://postgres:password@localhost:5432/gator?sslmode=disable",
    "current_user_name": ""
}
```

## Using gator
Install gator using `go install github.com/Cmolloy36/gator@latest`.

Commands:
* `addfeed`: Add new feed to collect from
* `agg`: Run in background, collects and creates posts from feeds
* `browse`: Browse posts
* `feeds`: View all feeds
* `follow`: Follow a previously unfollowed feed on current user
* `following`: See a list of all feeds the current user follows
* `login`: Log in as existing user (no authN yet!)
* `register`: Register a new user
* `reset`: Reset the DB
* `unfollow`: Unfollow a previously followed feed on current user
* `users`: See list of users, including current logged in user


## Future Ideas
- [ ] Add a help command that explains the commands available to the user
- [ ] Add sorting and filtering options to the browse command
- [ ] Add pagination to the browse command
- [ ] Add concurrency to the agg command so that it can fetch more frequently
- [ ] Add a search command that allows for fuzzy searching of posts
- [ ] Add bookmarking or liking posts
- [ ] Add a TUI that allows you to select a post in the terminal and view it in a more readable format (either in the terminal or open in a browser)
- [ ] Add an HTTP API (and authentication/authorization) that allows other users to interact with the service remotely
- [ ] Write a service manager that keeps the agg command running in the background and restarts it if it crashes