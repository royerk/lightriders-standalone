#!/usr/bin/env python3

from math import cos, pi, sin, sqrt, atan
import sys

import random

from fractions import Fraction
import operator
from game import Game
from copy import deepcopy

import player

try:
    from sys import maxint
except ImportError:
    from sys import maxsize as maxint

LAND = -2
WATER = -1
MAP_OBJECT = '.%'

PLAYER0 = 0
PLAYER1 = 1
WALL = 2
EMPTY = 3

VALID_ORDERS = ["up", "down", "left", "right", "pass"]

ADJACENT = [
    (-1, 0),
    (0, 1),
    (1, 0),
    (0, -1)
]

CODE_SPAWN_DELAY = 8
WEAPON_SPAWN_DELAY = 20
BUG_SPAWN_COUNT = 5
HIT_PENALTY = 4

class Lightriders(Game):
    def __init__(self, options=None):
        self.width = 0
        self.height = 0
        self.cutoff = None
#        self.turns = int(options['turns'])
#        self.loadtime = int(options['loadtime'])
#        self.turntime = int(options['turntime'])
        self.engine_seed = options.get('engine_seed',
            random.randint(-maxint-1, maxint))
        random.seed(self.engine_seed)
        self.timebank = 0
        if 'timebank' in options:
            self.timebank = int(options['timebank'])
        self.time_per_move = int(options['time_per_move'])
        self.player_names = ["player0", "player1"] #options['player_names']
        self.player_seed = options.get('player_seed',
            random.randint(-maxint-1, maxint))

        self.turn = 0
        self.turn_limit = int(options['turns'])
        self.num_players = 2 # map_data["players"]
        self.players = [player.Player(), player.Player()]
        # used to cutoff games early
        self.cutoff_turns = 0
        # used to calculate the turn when the winner took the lead
        self.winning_bot = None
        self.winning_turn = 0
        # used to calculate when the player rank last changed
        self.ranking_bots = None
        self.ranking_turn = 0

        # initialize scores
        self.score = [0]*self.num_players
        self.bonus = [0]*self.num_players
        self.score_history = [[s] for s in self.score]

        # used to track dead players
        self.killed = [False for _ in range(self.num_players)]

        # used to give a different ordering of players to each player;
        # initialized to ensure that each player thinks they are player 0
        # Not used in this game I think.
        self.switch = [[None]*self.num_players + list(range(-5,0))
                       for i in range(self.num_players)]
        for i in range(self.num_players):
            self.switch[i][i] = 0

        # the engine may kill players before the game starts and this is needed
        # to prevent errors
        self.orders = [[] for i in range(self.num_players)]

        ### collect turns for the replay
        self.replay_data = []

        self.map_data = self.new_map ()


    def string_cell_item(self, item):
        if item == PLAYER0:
            return '0'
        elif item == PLAYER1:
            return '1'
        elif item == EMPTY:
            return '.'
        elif item == WALL:
            return 'x'
        else:
            return ' '

    def in_bounds (self, row, col):
        return row >= 0 and col >= 0 and col < self.width and row < self.height

    def output_cell (self, cell):
        return self.string_cell_item(cell)

    def string_field (self, field):
        flat = []
        for row in field:
            for cell in row:
                flat.append(cell)
        return (','.join([self.output_cell (cell) for cell in flat]))

    def init_grid (self, rows, cols):

        grid = []

        for row in range(0, rows):
            grid.append([])
            for col in range(0, cols):
                grid[row].append(EMPTY)

        return grid

    def choose_player_locs (self, rows, cols):
        row = random.randint(0, rows - 1)
        col_offset = random.randint(1, (cols - 2) / 2)
        self.players[0].row = row
        self.players[0].prev_row = row
        self.players[1].row = row
        self.players[1].prev_row = row

        self.players[0].col = col_offset
        self.players[0].prev_col = col_offset
        self.players[1].col = cols - (col_offset + 1)
        self.players[1].prev_col = cols - (col_offset + 1)

        self.field[row][col_offset] = PLAYER0
        self.field[row][cols - (col_offset + 1)] = PLAYER1

    def new_map (self):
        rows = 16#random.randint(10, 20)
        cols = 16#random.randint(10, 20)
        self.height = rows
        self.width = cols
        grid = self.init_grid (rows, cols)
        self.field = grid
        self.choose_player_locs(rows, cols)
        return {
            "size": (rows, cols),
            "grid" : grid }

    def render_changes(self, player, time_to_move):
        """ Create a string which communicates the updates to the state
        """
        updates = self.get_state_changes(time_to_move)
        visible_updates = []
        # next list all transient objects
        for update in updates:
            visible_updates.append(update)
        visible_updates.append([]) # newline
        return '\n'.join(' '.join(map(str,s)) for s in visible_updates)

    def get_state_changes(self, time_to_move):
        """ Return a list of all transient objects on the map.

        """
        changes = []
        changes.extend([['update game round', self.turn + 1]])
        changes.extend([['update game field', self.string_field(self.field)]])
        changes.extend([['action move', int(time_to_move * 1000)]]) 
        return changes

    def convert_move(self, move):
        """ Convert text string to co-ordinate offset
        """
        if move == "up":
            return {"row" : -1, "col" : 0}
        elif move == "down":
            return {"row": 1, "col" : 0}
        elif move == "left":
            return {"row": 0, "col" : -1}
        elif move == "right":
            return {"row": 0, "col" : 1}
        elif move == "pass":
            return None
        else:
            raise ValueError("Failed to convert string to move: " + move)

    def parse_orders(self, player, lines):
        """ Parse orders from the given player
        player is an integer
        """
        orders = []
        valid = []
        ignored = []
        invalid = []

        for line in lines:
            line = line.strip().lower()
            # ignore blank lines and comments
            if not line: # or line[0] == '#':
                continue

            if line[0] == '#':
                ignored.append((line))
                continue

            data = line.split()

            try:

                # validate data format
                if data[0] not in VALID_ORDERS:
                    invalid.append((line, 'unknown action'))
                    continue
                else:
                    move = self.convert_move(data[0])
                if move:
                    row, col = self.players[player].row + move['row'], self.players[player].col + move['col']
                    orders.append((player, row, col))

                valid.append(line)

            except ValueError:
                invalid.append((line, "submitted move is invalid: " + line))
                continue

        return orders, valid, ignored, invalid



    def bots_to_play(self, turn):
        """ Indices of the bots who should receive the game state and return orders now """
        return [0, 1]

    def board_symbol(self, cell):
        if PLAYER0 == cell:
            return '0'
        elif PLAYER1 == cell:
            return '1'
        elif EMPTY == cell:
            return '.'
        elif WALL == cell:
            return 'x'
        else:
            raise ValueError ("Invalid board_symbol: " + str(cell))

    def text_board(self):
        print ("\n")
        for row in self.field:
            print("")
            for cell in row:
                sys.stdout.write(self.board_symbol(cell))
        print("\n")

    def player_cell (self, player):
        if player == 0: return PLAYER0 
        else: return PLAYER1

    def adjacent_coords (self, row, col):
        result = []
        for (orow, ocol) in ADJACENT:
            (trow, tcol) = (row + orow, col + ocol)
            if self.in_bounds (trow, tcol):
                result.append((trow, tcol))
        return result

    def not_blocked(self, row, col):
        return (self.field[row][col] != WALL)

    def is_legal(self, player, row, col):
        in_range = (row, col) in self.adjacent_coords(self.players[player].row, self.players[player].col)
        not_reverse = (row != self.players[player].prev_row or col != self.players[player].prev_col)
        return in_range and not_reverse and self.not_blocked (row, col)

    def place_move(self, move):
        (player, row, col) = move
        p = self.players[player]
        if self.is_legal (player, row, col):
            self.field[p.row][p.col] = WALL
            self.field[row][col] = player
            p.prev_row = p.row
            p.prev_col = p.col
            p.row = row
            p.col = col
        else:
            self.kill_player (player)

    def check_collide_heads(self):
        if (self.players[0].row == self.players[1].row) and (self.players[0].col == self.players[1].col):
            self.kill_player(0)
            self.kill_player(1)

    def do_orders(self):
        """ Execute player orders and handle conflicts
        """
        for player in self.bots_to_play(self.turn):
            if self.is_alive(player):
                if len(self.orders[player]) > 0:
                    self.place_move (self.orders[player][0])
            else:
                pass
        self.check_collide_heads()

    # Common functions for all games

    def game_over(self):
        """ Determine if the game is over

            Used by the engine to determine when to finish the game.
            A game is over when there are no players remaining, or a single
              winner remaining.
        """
        if len(self.remaining_players()) < 1:
            self.cutoff = 'extermination'
            return True
        elif len(self.remaining_players()) == 1:
            self.cutoff = 'lone survivor'
            return True
        elif self.turn >= self.turn_limit:
            self.cutoff = 'turn limit reached'
            return True
        else: return False

    def kill_player(self, player):
        """ Used by engine to signal that a player is out of the game """
        print("Player killed: " + str(player))
        self.killed[player] = True
        self.score[player] = -1

    def start_game(self):
        """ Called by engine at the start of the game """
        self.game_started = True
        
        ### append turn 0 to replay
        self.replay_data.append( self.get_state_changes(self.time_per_move) )
        result = []

    def score_game(self):
            return [self.score[0], self.score[1]]

    def finish_game(self):
        """ Called by engine at the end of the game """

        self.score = self.score_game()
#        self.text_board()
#        self.text_macroboard()
        self.calc_significant_turns()
        for i, s in enumerate(self.score):
            self.score_history[i].append(s)
        self.replay_data.append( self.get_state_changes(self.time_per_move) )

        # check if a rule change lengthens games needlessly
        if self.cutoff is None:
            self.cutoff = 'turn limit reached'

    def start_turn(self):
        """ Called by engine at the start of the turn """
        self.turn += 1
        self.text_board()
        self.orders = [[] for _ in range(self.num_players)]

    def finish_turn(self):
        """ Called by engine at the end of the turn """
        self.do_orders()
        # record score in score history
        for i, s in enumerate(self.score):
            if self.is_alive(i):
                self.score_history[i].append(s)
            elif s != self.score_history[i][-1]:
                # score has changed, increase history length to proper amount
                last_score = self.score_history[i][-1]
                score_len = len(self.score_history[i])
                self.score_history[i].extend([last_score]*(self.turn-score_len))
                self.score_history[i].append(s)
        self.calc_significant_turns()
#        self.update_scores()

        ### append turn to replay
        self.replay_data.append( self.get_state_changes(self.time_per_move) )

    def calc_significant_turns(self):
        ranking_bots = [sorted(self.score, reverse=True).index(x) for x in self.score]
        if self.ranking_bots != ranking_bots:
            self.ranking_turn = self.turn
        self.ranking_bots = ranking_bots

        winning_bot = [p for p in range(len(self.score)) if self.score[p] == max(self.score)]
        if self.winning_bot != winning_bot:
            self.winning_turn = self.turn
        self.winning_bot = winning_bot

    def get_state(self):
        """ Get all state changes

            Used by engine for streaming playback
        """
        updates = self.get_state_changes()
        updates.append([]) # newline
        return '\n'.join(' '.join(map(str,s)) for s in updates)

    def get_player_start(self, player=None):
        """ Get game parameters visible to players

            Used by engine to send bots startup info on turn 0
        """
        result = []
        result.append(['settings player_names', ','.join(self.player_names)])
        result.append(['settings your_bot', self.player_names[player]])
        result.append(['settings timebank', self.timebank])
        result.append(['settings time_per_move', self.time_per_move])
        result.append(['settings your_botid', player])
        result.append(['settings field_width', self.width])
        result.append(['settings field_height', self.height])
        result.append(['settings max_rounds', self.turn_limit])

        result.append(['settings player_seed', self.player_seed])
        #result.append(['settings num_players', self.num_players])
        #message = self.get_player_state(player, self.timebank)

        result.append([]) # newline
        pen = '\n'.join(' '.join(map(str,s)) for s in result)
        return pen #+ message #+ 'ready\n'

    def get_player_state(self, player, time_to_move):
        """ Get state changes visible to player

            Used by engine to send state to bots
        """
        return self.render_changes(player, time_to_move)

    def is_alive(self, player):
        """ Determine if player is still alive

            Used by engine to determine players still in the game
        """
        if self.killed[player]:
            return False
        else:
            return True

    def get_error(self, player):
        """ Returns the reason a player was killed

            Used by engine to report the error that kicked a player
              from the game
        """
        return ''

    def do_moves(self, player, moves):
        """ Called by engine to deliver latest player orders """
        orders, valid, ignored, invalid = self.parse_orders(player, moves)
#        orders, valid, ignored, invalid = self.validate_orders(player, orders, valid, ignored, invalid)
        self.orders[player] = orders
        return valid, ['%s # %s' % ignore for ignore in ignored], ['%s # %s' % error for error in invalid]

    def get_scores(self, player=None):
        """ Gets the scores of all players

            Used by engine for ranking
        """
        #if player is None:
        return [0 - s for s in self.score]
        #else:
        #    return self.order_for_player(player, self.score)

    def order_for_player(self, player, data):
        """ Orders a list of items for a players perspective of player #

            Used by engine for ending bot states
        """
        s = self.switch[player]
        return [None if i not in s else data[s.index(i)]
                for i in range(max(len(data),self.num_players))]

    def remaining_players(self):
        """ Return the players still alive """
        return [p for p in range(self.num_players) if self.is_alive(p)]

    def get_stats(self):
        """  Used by engine to report stats
        """
        stats = {"scores": [self.score[0], self.score[1]]}
        return stats

    def get_replay(self):
        """ Return a summary of the entire game

            Used by the engine to create a replay file which may be used
            to replay the game.
        """
        replay = {}
        # required params
        replay['revision'] = 1
        replay['players'] = self.num_players

        # optional params
        replay['loadtime'] = self.timebank
        replay['turntime'] = self.time_per_move
        replay['turns'] = self.turn
        replay['engine_seed'] = self.engine_seed
        replay['player_seed'] = self.player_seed

        # scores
        replay['scores'] = self.score_history
        replay['bonus'] = self.bonus
        replay['winning_turn'] = self.winning_turn
        replay['ranking_turn'] = self.ranking_turn
        replay['cutoff'] =  self.cutoff

        
        ### 
        replay['data'] = self.replay_data
        return replay


    def bot_input_finished(self, line):
        return line.strip().lower() in VALID_ORDERS

