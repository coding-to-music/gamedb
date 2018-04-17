#!/usr/bin/env bash

if [ "${ENV}" == "local" ]
then

#    gcloud config set project ${STEAM_GOOGLE_PROJECT}
#    $(gcloud beta emulators datastore env-init)
#    gcloud beta emulators datastore start &

    bash ./cmd/pics.sh >> /dev/null
    realize start

else

    steam-authority --pics --consumers

fi
