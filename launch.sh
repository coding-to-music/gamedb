#!/bin/bash

export HOME=$(pwd)

mkdir -p ./logs
touch ./logs/gamedb.log
touch ./logs/grunt.log
touch ./logs/updater.log

grunt >>${HOME}/logs/grunt.log 2>&1 &
export GRUNT_PID=$!
echo "Loaded Grunt"

cd /Websites/steam-authority/steam-updater/ && dotnet run >>${HOME}/logs/updater.log 2>&1 &
export UPDATER_PID=$!
echo "Loaded Updater"

realize start
#export REALIZE_PID=$!
#echo "Loaded Game DB"

#lnav ${HOME}/logs/

function finish {
    echo "Killing everything"

#    kill ${REALIZE_PID}
    kill ${GRUNT_PID}
    kill ${UPDATER_PID}

    # Kill dotnet's spawned processes
    kill $(ps aux | grep 'SteamUpdater.dll' | awk '{print $2}') >>/dev/null 2>&1
}

trap finish EXIT

wait

#while true; do
#    read -rsn1 input
#    if [ "$input" = "q" ]; then
#        exit
#    fi
#done
