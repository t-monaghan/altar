# Summary

The `github/contributions` package defines a fetcher and handler for displaying a user's GitHub contribution graph. This is done by utilising the GitHub cli extension `gh-altar`.

## Usage

### Setup

The setup for an altar broker to utilise this example can be found [here](https://github.com/t-monaghan/altar/blob/main/main.go). You will also need to install the `gh-altar` extension to `gh` by running `gh extension install t-monaghan/gh-altar`. This extension is required to query the contribution graph for the given user.

### Running

Once your broker is running run `gh altar contributions --broker-address {YOUR_BROKERS_ADDRESS_HERE}`.

## Links

- [gh-altar's repository](https://github.com/t-monaghan/gh-altar)
