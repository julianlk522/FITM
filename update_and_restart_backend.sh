#!/bin/bash

LOG_FILE="/var/log/fitm_update.log"
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# redirect non-explicit output to log file (append)
# "exec >> {arg}" replaces current shell process (modifying stdout file descriptor) with {arg} output for later script commands
# "2>&1" redirects stderr (file descriptor 2) to stdout (1)
exec >> "$LOG_FILE" 2>&1

log "UPDATE AND RESTART"
echo

if [ -z "$FITM_ROOT_PATH" ]; then
    log "error: FITM_ROOT_PATH is not set"
    exit 1
fi

# navigate to root
cd "$FITM_ROOT_PATH" || { log "error: could not navigate to $FITM_ROOT_PATH"; exit 1; }
log "navigated to $FITM_ROOT_PATH"

# pull changes
git pull

# navigate to backend
if [ ! -d "backend" ]; then
    log "error: 'backend' directory not found"
    exit 1
fi
cd backend
log "navigated to /backend"


# update dependencies, rebuild
go mod tidy
go build --tags 'fts5' .
log "build complete"

# get running server process ID(s)
PIDs=$(pgrep -f fitm)
log "found PID(s): $PIDs"

# interrupt if running
if [ -n "$PIDs" ]; then
    for PID in $PIDs; do
        log "attempting to stop process $PID"
        # send SIGTERM signal to gracefully stop process
        kill $PID
        
        # countdown process stop
        countdown=10

        # while process exists
        ## (kill -0 evals to status 0 if process exists and 1 if process does not exist)
        ## (2>/dev/null redirects stderr to null device file to suppress)
        while kill -0 $PID 2>/dev/null; do
            if [ $countdown -le 0 ]; then
                log "countdown exceeded for PID $PID. Forcing kill."
                kill -9 $PID
                break
            fi
            sleep 1
            ((countdown--))
        done
        log "stopped process $PID"
    done
fi
log "all old processes stopped"

# start tmux session if not exists already
if ! tmux has-session -t FITM 2>/dev/null; then
    log "creating new FITM tmux session"
    tmux new-session -d -s FITM
fi

# start new binary in tmux session
tmux send-keys -t FITM "cd $FITM_ROOT_PATH/backend && ./fitm" ENTER

# detach
tmux detach -s FITM

# save timestamp
log "update complete and server restarted"