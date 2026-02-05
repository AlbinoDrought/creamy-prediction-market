# Creamy Prediction Market

<p align="center">
    <img src="./.readme/logo.png" alt="Creamy Prediction Market Cat Logo">
    <p align="center">The creamiest selfhosted prediction market</p>
    <p align="center">
        <a href="https://github.com/AlbinoDrought/creamy-prediction-market/blob/master/LICENSE"><img alt="AGPL-3.0 License" src="https://img.shields.io/github/license/AlbinoDrought/creamy-prediction-market"></a>
    </p>
</p>

This application is for hosting a fun prediction market game with friends. 

As an example, use this during a party to have people submit predictions and have others bet fake tokens on them.

This is not intended for production use. This application is unlikely to function correctly.

## Usage

Create a config like this:

```json
{
  "debug": false,
  "admin_pin": "1234",
  "repo_path": "/path/to/dbfile.json",
  "starting_tokens": 1000
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
