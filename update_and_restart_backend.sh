#!/bin/bash

if [ -z "$FITM_ROOT_PATH" ]; then
    echo "Error: FITM_ROOT_PATH is not set"
    exit 1
fi

# navigate to root
cd "$FITM_ROOT_PATH" || { echo "Error: could not navigate to $FITM_ROOT_PATH"; exit 1; }
echo "navigated to $FITM_ROOT_PATH"

# pull changes
git pull

# navigate to backend
if [ ! -d "backend" ]; then
    echo "Error: 'backend' directory not found"
    exit 1
fi
cd backend

# update dependencies, rebuild
go mod tidy
go build --tags 'fts5' .

# get running server process ID
PID=$(pgrep -f fitm)
# interrupt if running
if [ -n "$PID" ]; then

    # send SIGTERM signal to gracefully stop process
    kill $PID
    
    # countdown process stop
    countdown=10

    # while process exists
    ## (kill -0 evals to status 0 if process exists and 1 if process does not exist)
    ## (2>/dev/null redirects stderr to null device file to suppress)
    while kill -0 $PID 2>/dev/null; do
        if [ $countdown -le 0 ]; then
            echo "Countdown exceeded. Forcing kill."
            # send SIGKILL if needed
            kill -9 $PID
            break
        fi
        sleep 1
        ((countdown--))
    done
fi

# start tmux session if not exists already
if ! tmux has-session -t FITM 2>/dev/null; then
    echo "Creating new FITM tmux session"
    tmux new-session -d -s FITM
fi

# start new binary in tmux session
tmux send-keys -t FITM "cd $FITM_ROOT_PATH/backend && ./fitm" ENTER

# detach
tmux detach -s FITM

# timestamp
echo "Update complete and server restarted ($(date))"