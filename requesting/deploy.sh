echo -n "BASE_URL(https://GCS or Cloud Run URL) ?> "
read BASE_URL
echo gcloud run deploy --source=. --region=us-central1 --set-env-vars=BASE_URL=$BASE_URL --allow-unauthenticated delivering
