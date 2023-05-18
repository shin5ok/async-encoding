# Cut each movie asynchronously according to requests and distribute it

## Procedure to setup the whole system
### 0. Prepare Google Cloud Project and environment variables
Sign in to your project,
```
gcloud auth login --update-adc
gcloud config set project <your project id>
```
And then,
```
export TF_VAR_domain=<your domain name>
export TF_VAR_region=us-central1
export TF_VAR_zone=us-central1-a
export TF_VAR_gcs=<your bucket name>
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

### 3. Build applications and deploy them to Cloud Run.  
- **requesting application** that is to accept request from each user.
```
make requesting
```
- **delivering application** that is to deliver movie to each appropriate user.
```
make delivering
```

### 4. (*Temporary*) Prepare GCS Proxy
Prepare gcs-proxy to deliver objects that be protected in GCS to external user securely.  
Find your GCS bucket name that will store your movie in advance.  
eg: shingo-bucket-movie-xxxxx
```
export GCS_BUCKET=<your bucketname>
```

Deploy it.
```
git clone https://github.com/shin5ok/gcs-proxy.git
cd gcs-proxy/
bash deploy.sh
```

### 5. Prepare to test to work
- Prepare some test movies as MPEG4 format.  
Over 1 mins movie would be nice.  

*Note: You must make sure where you can use the contents.*  
*DO NOT use ones you don't have right to edit.*

- Transfer them to GCS bucket you prepared in advance.   
If you use gcloud cli,
```
gcloud storage cp *.mp4 gs://<your bucketname>/
```

- List your transfered movies's name into config named movies.yaml  
Like this,
```
movie-1.mp4
20320301.mp4
foobar.mp4
```

###  6. Test it
- Make a lot of requests.
```
cd requesting-clients/
bash ./test-publish.sh 10000
```
This is an example to send 10000 messages as request contains source image ans cutting time range randomly.


- See Cloud Logging, that will show how processing runs.  
And then, you see the progress in Firestore and GCS.

- Open the site url that was assigned to **delivering** Cloud Run.

