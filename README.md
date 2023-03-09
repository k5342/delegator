# delegator
A simple discord bot to _delegate_ everything using discord slash commands based on your pre-defined profiles.

## Use cases
- You can _delegete_ to the bot to restart a game server process
- You can _delegate_ to the bot to backup your home server
- You can _delegate_ to the bot to run a pre-defined shell script 

## Install
(required) go 1.19, tested on Ubuntu 22.04 on WSL2  
(required) you need to issue a bot token at the Discord Developer Portal (<https://discord.com/developers/applications>)

1. Clone this repository: `git clone https://github.com/k5342/delegator`
1. Build this project: `make`
1. (You can find a binary named "delegator" if the build was succeed)
1. Setup config file: `./delegator init`
1. (You can find a configuration file named "config.yaml"; please edit the file: `$EDITOR config.yaml`)
1. Launch the bot: `./delegator run`

## Configuration
Delegator reads a configuration file formatted in YAML.

## Example

### Configuration
```yaml

# issue your discord bot token from Discord Developer Portal
discord_bot_token: ...

# you can list commands here
commands:
- name: date
  exec: /usr/bin/date # a full path to the command
  description: Returns a result of date command
- name: timeout
  exec: /usr/bin/sleep 10
  description: an example for execution time timeout
  timeout:
    seconds: 5 # default is infinity (= 0). this limits to 5s

```

### Screenshots
![success](https://user-images.githubusercontent.com/1993005/224086658-8f4c8d12-a2fc-4652-bf9f-060c8845134d.gif)
![timeout](https://user-images.githubusercontent.com/1993005/224086664-f03fede7-86b0-4f7d-ab70-9611f7c5942b.gif)
