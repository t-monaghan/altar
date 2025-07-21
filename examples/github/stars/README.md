# Summary

The `github/stars` package defines a fetcher and handler for receiving webhook requests from GitHub when a user stars a repository.


## Setup

The setup for an altar broker to utilise this example can be found [here](https://github.com/t-monaghan/altar/blob/main/main.go). You will also need to set up a [webhook from your GitHub repository](https://docs.github.com/en/webhooks/using-webhooks/creating-webhooks) to send requests to your awtrix broker with a [shared secret](https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries#creating-a-secret-token).
