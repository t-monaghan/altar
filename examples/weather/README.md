# Weather

Currently the weather examples contains the precipitation application. This app will present the time until the next chance of precipitation and what that chance of precipitation is.

## Usage

You can find an example usage of the precipitation package in the main file [here](https://github.com/t-monaghan/altar/blob/main/main.go). Note that this app requires the environment variables `LATITUDE` and `LONGITUDE` which will define the location that this app will monitor the precipitation forecast for. These values should be signed floats, such as LATITUDE="-37.814" rather than using cardinal directions.

For an example, the coordinates for Melbourne, Australia can be found in `.env.example`.
