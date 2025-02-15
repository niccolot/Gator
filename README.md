# Gator

Gator is a blog aggregator CLI app that let you follow, store locally and read rss feeds.



## How to setup

```sh
git clone https://github.com/niccolot/Gator
cd Gator
```

Install dependencies with

```sh
cd scripts
chmod +x install_dependencies.sh
./install_dependencies.sh
```

And set the database url in the `.env` file by running

```sh
cd scripts
chmod +x create_env_vars.sh
./create_env_vars.sh
cd ..
source .env
```

## Database 

The database is implemented using [postgres](https://www.postgresql.org/) with Go code generated by [sqlc](https://sqlc.dev/) and the database migrations handled by [goose](https://github.com/pressly/goose).

### Setup database

Run in the terminal
```sh
sudo -u postgres psql
```

Then create the sql database
```sql
CREATE DATABASE gator;
\c gator
ALTER USER postgres PASSWORD 'postgres';
```

## How to use

In order to build and start the application one can use the provided script

```sh
cd scripts
chmod +x start.sh
./start.sh
```
Or in alternative

```sh
go install github.com/niccolot/Gator@v1.0.0
```

(make sure you have properly set up the Go environment):

```sh
export PATH=$(go env GOPATH)/bin:$PATH
```

### Commands

By starting the applicatio you are prompted by the CLI interface

```
Gator >
```

* The `help` command will list all the avalable commands

* The usual workflow is: registering as user, adding some feeds to the database, starting the aggregation and browsing the latest feeds, eventually bookmarking or opening your favourite posts.

#### Users

The app is usable by different users. By default the first to register is set as **superuser** which as some privileges such as resetting the database and changing user's passwords.

Different users can be registered and logged in with the `login <username>`.

#### Feeds

In order to follow a feed one has to save it in the database with the `addfeed <feed url>` command. Other users can follow feeds already saved in the database with the `follow <feed url>` command.

#### Aggregation

By running the `aggregate <time between updates> [optional] -log` a background goroutine is called to fetch all the feeds concurrently and update the posts list. With the optional tag the aggreagation is logged in a `aggreagation.log` file in case one wants to check if something is going wrong. The aggreagation can be stopped anytime with the `stopagg` command.

#### Posts

An user can see the latest posts from the feeds he follows by running the `browse <num posts to show>` command, bookmark some of them or open them in the browser.