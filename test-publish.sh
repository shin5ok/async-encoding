
while :
do
    if which uuid > /dev/null;
    then
        uuid=`uuid`
    else
        uuid=`uuidgen`
    fi
    start=$((RANDOM % 50))
    end=$start+5
    message="{\"src\":\"test-1.mp4\",\"dst\":\"\",\"user_id\":\"${uuid}\",\"start\":$((start)),\"end\":$((end))}"
    echo $message
    gcloud pubsub topics publish --message=$message test
done
