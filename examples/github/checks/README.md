# Summary

The `github/checks` package defines a fetcher and handler for keeping an eye on a GitHub pull request's check run. This is done by utilising the GitHub cli extension `gh-altar`.

## Usage

### Setup

The setup for an altar broker to utilise this example can be found [here](https://github.com/t-monaghan/altar/blob/main/main.go). You will also need to install the `gh-altar` extension to `gh` by running `gh extension install t-monaghan/gh-altar`. This extension is required to query the pull request information for the working directory.

### Running

Once your broker is running and your terminal is open in a repository that has a pull request check run working away on GitHub you can begin polling the state and sending it to your broker by running `gh altar watch-ci --broker-address {YOUR_BROKERS_ADDRESS_HERE}`. The extension will begin polling the check run until the checks are complete, however it is safe to kill the extension if you wish to stop it before the checks finish.

## Links

- [gh-altar's repository](https://github.com/t-monaghan/gh-altar)
