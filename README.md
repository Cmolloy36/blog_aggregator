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
?
