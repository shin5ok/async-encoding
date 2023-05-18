TOPIC=test
gcloud run deploy --source=. --region=us-central1 --set-env-vars=TOPIC=$TOPIC --allow-unauthenticated requesting
