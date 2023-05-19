# Processing each movie asynchronously according to client requests and distribute it

## Prerequiste
- Google Cloud Account enabled Google Cloud billing
- Google Cloud Project

## Procedure to setup the whole system
### 1. Prepare Google Cloud Project and environment variables
Sign in to your project,
```
gcloud auth login --update-adc
gcloud config set project <your project id>
```
And then, set some environment variables to be used in the following step.
```
export TF_VAR_domain=<your domain name>
export TF_VAR_region=<GCP region ex: us-central1>
export TF_VAR_zone=<GCP region ex:us-central1-a>
export TF_VAR_gcs=<your bucket name>
export GOOGLE_CLOUD_PROJECT=<your project id>
```
Note: Domain name must be one you manage.  
If you don't have it, You can get your favorite one at *Cloud Domains*.

### 2. Build infrastructure
Preparation and confimation to run totally.
```
cd terraform/
rm *tfstate*
terraform plan
```
Just type it to build the whole infrastructure.
```
terraform apply
```

### 3. (*Temporary*) Prepare GCS Proxy
Prepare gcs-proxy to deliver objects that be protected in GCS to external user securely.  
Set environment variables of your bucket name from terraform's one.
```
export GCS_BUCKET=$TF_VAR_gcs
```

Deploy it.
```
git clone https://github.com/shin5ok/gcs-proxy.git
cd gcs-proxy/
export REGION=$TF_VAR_region
bash deploy.sh
```

Set environment variables of url the Cloud Run published for after procedure.
```
export BASE_URL=https://xxxxxxxxxxxx.run.app
```

### 4. Build applications and deploy them to Cloud Run.  
Before doing here, make sure if you have Docker environment on your PC, Or you need to prepare it.

- Prepare your registory for containers, you need to do it just only once.
```
make repo
```

- **requesting application** that is to accept request from each user.
```
make requesting
```
- **delivering application** that is to deliver movie to each appropriate user.
```
make delivering
```

### 5. Prepare to test to work
- Prepare some test movies as MPEG4 format.  
Over 1 mins movie would be nice.  

*Note: You must make sure where you can use the contents.*  
*DO NOT use ones you don't have right to edit.*

- Transfer them to GCS bucket you prepared in advance.   
If you use gcloud cli,
```
gcloud storage cp *.mp4 gs://$TF_VAR_gcs/
```

- List your transfered movies's name into config named movies.yaml  
Like this,
```
movie-1.mp4
20320301.mp4
foobar.mp4
```
You will make it simply by hit command in the right directory,
```
ls *.mp4 > YOUR_DIRECTORY/movies.txt
```

###  6. Test it
- Make a lot of requests.
Build a command to test.  
If you don't have 'go' command, install the latest one according to [here](https://go.dev/doc/install).
```
make test-client
```
Do test.
```
cd request-clients/
POST_URL=<your Cloud Run 'requesting' URL>
./test-client -posturl=$POST_URL -procnum 10 -requestnum 10000
```
This is an example to send 10000 messages as request contains source image and cutting time range randomly, from 10 virtual clients parallelly.


- See Cloud Logging, that will show how processing runs.  
And then, you also see the progress in Firestore and GCS.

- Open the site url that was assigned to **delivering** Cloud Run with your browser.

## 7. Cleanup
See [here](https://cloud.google.com/resource-manager/docs/creating-managing-projects?shutting_down_projects&hl=ja#shutting_down_projects).
