# cinemateque-ics

Parse the HTML of [cinemateket.dk](https://cinemateket.dk) and output an `.ics` readable in Outlook, Google Calendar, iCal or the like.

## Build

### Binary
`make build`

## Lambda function
`make build-lambda`

Upload .zip to Lambda, configure an API gateway

## Run

### Single URL

`./build/cin2ics -url https://www.dfi.dk/cinemateket/biograf/alle-film/film/big-blue`

## Multiple in one file

`./build/cin2ics -file path-to-newline-seperated-file`

## Import

`open events.ics`
