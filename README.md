# Creamy Prediction Market

<p align="center">
    <img src="./.readme/logo.png" alt="Creamy Prediction Market Cat Logo">
    <p align="center">The creamiest selfhosted prediction market</p>
    <p align="center">
        <a href="https://albinodrought.github.io/creamy-prediction-market/"><img alt="Website" src="https://img.shields.io/website?url=https%3A%2F%2Falbinodrought.github.io%2Fcreamy-prediction-market%2F&label=website"></a>
        <a href="https://github.com/AlbinoDrought/creamy-prediction-market/blob/master/LICENSE"><img alt="AGPL-3.0 License" src="https://img.shields.io/github/license/AlbinoDrought/creamy-prediction-market"></a>
    </p>
</p>

This application is for hosting a fun prediction market game with friends. 

As an example, use this during a party to have people submit predictions and have others bet fake tokens on them.

This is not intended for production use. This application is unlikely to function correctly.

## Screenshots

| Screenshot | Description |
|---|---|
| <img src=".readme/Screen Shot 2026-02-06 at 13.33.21.png" width="250" /> | **Login**: Simple name + PIN login screen |
| <img src=".readme/Screen Shot 2026-02-06 at 13.33.42.png" width="250" /> | **Home (empty)**: Home screen with no predictions yet |
| <img src=".readme/Screen Shot 2026-02-06 at 13.37.23.png" width="250" /> | **Home (active)**: An open prediction card on the home feed |
| <img src=".readme/Screen Shot 2026-02-06 at 13.37.27.png" width="250" /> | **Prediction detail**: Viewing a prediction with choices and live odds |
| <img src=".readme/Screen Shot 2026-02-06 at 13.37.31.png" width="250" /> | **Placing a bet (10 tokens)**: Bet amount selector with a choice selected |
| <img src=".readme/Screen Shot 2026-02-06 at 13.37.35.png" width="250" /> | **Placing a bet (100 tokens)**: Larger bet amount with updated potential payout |
| <img src=".readme/Screen Shot 2026-02-06 at 13.37.38.png" width="250" /> | **Achievement unlocked**: "First Bet" achievement popup after placing a bet |
| <img src=".readme/Screen Shot 2026-02-06 at 13.37.42.png" width="250" /> | **Bet placed**: Prediction view showing the user's active bet and current odds |
| <img src=".readme/Screen Shot 2026-02-06 at 13.37.54.png" width="250" /> | **Prediction closed**: Betting is closed, waiting for the outcome to be decided |
| <img src=".readme/Screen Shot 2026-02-06 at 13.38.00.png" width="250" /> | **Achievement unlocked**: "Double Up" achievement for winning 2x a bet |
| <img src=".readme/Screen Shot 2026-02-06 at 13.40.16.png" width="250" /> | **Prediction decided**: Outcome revealed with winning choice highlighted in green and payout shown |
| <img src=".readme/Screen Shot 2026-02-06 at 13.40.22.png" width="250" /> | **Leaderboard**: Player rankings with avatars, cosmetics, titles, and achievement badges |
| <img src=".readme/Screen Shot 2026-02-06 at 13.40.30.png" width="250" /> | **My Bets**: Personal bet history with win/loss stats and net result |
| <img src=".readme/Screen Shot 2026-02-06 at 13.40.33.png" width="250" /> | **Shop**: Item shop with hats, effects, emojis, and other cosmetics for sale |
| <img src=".readme/Screen Shot 2026-02-06 at 13.40.39.png" width="250" /> | **Shop purchase modal**: Preview of a Crown hat on the user's avatar before buying |
| <img src=".readme/Screen Shot 2026-02-06 at 13.40.42.png" width="250" /> | **Item equipped**: Crown hat purchased and equipped, shown on avatar in the header |
| <img src=".readme/Screen Shot 2026-02-06 at 13.41.03.png" width="250" /> | **Name effect preview**: Rainbow name effect shown in the purchase preview modal |
| <img src=".readme/Screen Shot 2026-02-06 at 13.41.21.png" width="250" /> | **Dino minigame**: Chrome dino-style runner game for earning shop coins |
| <img src=".readme/Screen Shot 2026-02-06 at 13.41.29.png" width="250" /> | **Minigame achievement**: "Gamer" achievement unlocked after first minigame play |
| <img src=".readme/Screen Shot 2026-02-06 at 13.41.31.png" width="250" /> | **Minigame game over**: Score below 100, no coins earned |
| <img src=".readme/Screen Shot 2026-02-06 at 13.42.13.png" width="250" /> | **Minigame coins earned**: Score of 105 earns +1 coin for the shop |
| <img src=".readme/Screen Shot 2026-02-06 at 13.42.20.png" width="250" /> | **Leaderboard (customized)**: Players with full cosmetics: hats, name effects, titles, and avatar items |

## Usage

Create a config like this:

```json
{
  "debug": false,
  "admin_pin": "1234",
  "repo_path": "/path/to/dbfile.json",
  "starting_tokens": 1000,
  "starting_coins": 5
}
```

Give it to creamy-prediction-market by saving it to `config.json` or via env in `CREAMY_PM_CONFIG`

Build the project with `make`

Run the project with `./creamy-prediction-market`

View the project at http://localhost:3000

### Container

```
docker build -t creamy-prediction-market .
docker run \
  --rm \
  -p 3000:3000 \
  creamy-prediction-market
```
