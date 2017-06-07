LightRiders Engine
=========================

Standalone engine for LightRiders on starapple.riddles.io
---------------------------------------------------------

In development, there may be bugs!

Timekeeping is not enforced properly at this point: it simply gives a fixed amount of time per turn.

Grid size is generated randomly to improve testing capacities, for bots with hard coded grid size check the *new_map()* function in *lightriders.py* (line ~154).

You can use the play_one_game scripts to run single matches.

You can also use the manager.py script for automated testing like so:

    python3 manager.py -A bot1 -p "path or command to run bot1"
    python3 manager.py -A bot2 -p "path or command to run bot2"
    python3 manager.py -A bot3 -p "path or command to run bot3"
    python3 manager.py -f

The first three lines add some bots, and the last line will run matches forever, printing a ranking table after every match until interrupted by a keypress.

You can see the list of command line options by running manager.py without any arguments:

    python3 manager.py

The playgame.py module has its own list of command line arguments, but to inspect those you have to look at the source.

The manager utility relies on the "skills" module, which can be installed through pip:

    pip3 install skills

Good luck!

##### Erratum for Go bots
If you have a Go bot made of multiple files which contains classes you won't be able to use `go run /path/*.go` as the bot command. You need to concatenate your code into a single file, with tools like [GoConcatenator](https://github.com/royerk/Python-GoConcatenator) for example. Then the command `go run /path/golangBlob.go` is a valid command for the manager.
