
while :
do
    if which uuid
    then
        uuid=`uuid`
    else
        uuid=`uuidgen`
    fi
    message="{\"src\":\"test-1.mp4\",\"dst\":\"bar.mp4\",\"user_id\":\"${uuid}\",\"start\":200.0,\"end\":230.0}"
    echo $message
    gcloud pubsub topics publish --message=$message test
done
