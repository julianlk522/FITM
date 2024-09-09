#!/bin/sh

# pull changes
cd $FITM_ROOT_PATH
git pull

# update dependencies
cd backend
go mod tidy

# rebuild
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

# start new binary in tmux session
tmux send-keys -t FITM "./fitm" ENTER

# Detach from the tmux session
tmux detach -s FITM

# timestamp
echo "Update complete and server restarted ($(date))"