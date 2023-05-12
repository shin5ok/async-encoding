# METADATA_VALUE=$(curl http://metadata.google.internal/computeMetadata/v1/instance/attributes/foo -H "Metadata-Flavor: Google")
GOOGLE_CLOUD_PROJECT=$(curl http://metadata.google.internal/computeMetadata/v1/project/project-id -H "Metadata-Flavor: Google")
BUCKET=$GOOGLE_CLOUD_PROJECT
SUBSCRIPTION=$(curl http://metadata.google.internal/computeMetadata/v1/instance/attributes/SUBSCRIPTION -H "Metadata-Flavor: Google")

curl -sSO https://dl.google.com/cloudagents/add-google-cloud-ops-agent-repo.sh
bash add-google-cloud-ops-agent-repo.sh --also-install

apt update && apt -y install docker.io
sleep 5

gcloud auth configure-docker --quiet
sleep 1


docker run -d \
    --log-driver=gcplogs --log-opt=gcp-project=$GOOGLE_CLOUD_PROJECT \
    --restart always -e GOOGLE_CLOUD_PROJECT=$GOOGLE_CLOUD_PROJECT -e BUCKET=$BUCKET -e SUBSCRIPTION=$SUBSCRIPTION gcr.io/$GOOGLE_CLOUD_PROJECT/cutting:latest
