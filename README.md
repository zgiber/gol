# gol
Playing around with Game of Life 

## What?
It's an experimental small piece of code: a Game of life server and a simple browser client. 

## How?
Clone the repository, compile `go build` then run the binary. Once the server is running, open a browser an go to http://localhost:8080 (or another port if you specified it).
In the browser moving your cursor over the canvas creates "life". Move it slowly for a better effect. Anyone connected to the server will
see the same thing and can interact with it. Unfortunately it takes quite some bandwidth to sync when there is a lot going on, so it won't work with too many clients.

## What next?
I don't really have plans with this. Maybe I'll try to save bandwidth, or come up with a smarter sync.
