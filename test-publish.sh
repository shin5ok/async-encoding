n=${1:-10}
echo $n
if which uuid > /dev/null;
then
    uuid_cmd=uuid
else
    uuid_cmd=uuidgen
fi

while [ $((n)) -gt 0 ];
do
    echo "message $n"
    start=$((RANDOM % 50))
    end=$start+5
    uuid=`$uuid_cmd`
    message="{\"src\":\"test-1.mp4\",\"dst\":\"\",\"user_id\":\"${uuid}\",\"start\":$((start)),\"end\":$((end))}"
    echo $message
    gcloud pubsub topics publish --message=$message test
    n=$((n - 1))
done
