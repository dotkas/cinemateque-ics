# cinemateque-ics

Parse the HTML of [cinemateket.dk](https://cinemateket.dk) and output an `.ics` readable in Outlook, Google Calendar, iCal or the like.

## Build

`go build -o .build/cin2ics`

## Run

### Single URL

`./build/cin2ics -url https://www.dfi.dk/cinemateket/biograf/alle-film/film/big-blue`

## Multiple in one file

`./build/cin2ics -file path-to-newline-seperated-file`

## Import

`open events.ics`
