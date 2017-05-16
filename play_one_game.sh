#!/bin/sh

./playgame.py --verbose --fill --log_input --log_output --log_error \
--log_dir game_logs --map_file map.txt --log_stderr --turns 200 --turntime 500 --loadtime 1500 --nolaunch \
"./ocaml_starter" \
"./ocaml_starter"
