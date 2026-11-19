# Redis Clone

This is my implementation of a Redis clone using the Redis protocol to handle several commands

## Features

Currently the following functionality is implemented:

1. Basic commands: PING, ECHO, SET (including expiration behavior), GET, TYPE, INCR
2. Lists: Pushing, Removing and blocking retrieval.
3. Streams: Creation of streams, ID validations, querying, blocking reads.
4. Transactions: MULTI, DISCARD and EXEC commands.

The following features are in progress:

1. Replication
