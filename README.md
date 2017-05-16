LightRiders Engine
=========================

Standalone engine for LightRiders on starapple.riddles.io
---------------------------------------------------------

In development, there may be bugs!

Timekeeping is not enforced properly at this point: it simply gives a fixed amount of time per turn.

You can use the play_one_game scripts to run single matches.

You can also use the manager.py script for automated testing like so:

    python3 manager.py -A bot1 -p "path or command to run bot1"
    python3 manager.py -A bot2 -p "path or command to run bot2"
    python3 manager.py -A bot3 -p "path or command to run bot3"
    python3 manager.py -f

This will run matches forever, printing a ranking table after every match, untill interrupted by a keypress.

You can see the full list of command line options by running manager.py without any arguments:

    python3 manager.py

This utility relies on the "skills" module, which can be installed through pip:

    pip3 install skills

Good luck!
