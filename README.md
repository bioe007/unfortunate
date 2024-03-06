unfortunate
===========
Acts like the 'fortune' command where given a file (plain text) of quotes, pick
random one and print it to screen.

To do this quickly the first time it's run unfortunate will create a cache with
offsets to the fortune locations. Subsequently running `unfortunate` will output
a string like:
'''8:     eight'

where '8' is the random number and the text 'eight' is actually the fortune.
This is just so I could verify basic operation.

This is a little semi-working/hardly tested exercise in figuring out binary file
management in Go. There are lots of style issues here...


### Use

either build or just use `go run .`

There is a fake 'fortune' file included `./fakefortune.txt` which is the default
data file path. Then a unfortante.cache file will also be made in CWD.


### TODO
  - Should really hash the input file and store it in the cache
  - I did intend to store the lengths also but it's probably not worth it so
    this could be removed from the structure/file
  - clean up a lot of constants either un-used or undefined (i.e. '%')

