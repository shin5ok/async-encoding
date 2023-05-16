
n=${1:-10}

list=(`grep -v ^$ ../movies.txt|xargs`)
list_count=${#list[@]}

if which uuid > /dev/null;
then
    uuid_cmd=uuid
else
    uuid_cmd=uuidgen
fi

while [ $((n)) -gt 0 ];
do
    echo "message $n"
    start=$((RANDOM % 20))
    end=$start+5
    uuid=`$uuid_cmd`
    index=$((RANDOM % ${list_count}))
    src=${list[$((index))]}
    message="{\"src\":\"${src}\",\"dst\":\"\",\"user_id\":\"${uuid}\",\"start\":$((start)),\"end\":$((end))}"
    echo $message
    gcloud pubsub topics publish --message=$message test
    n=$((n - 1))
done
