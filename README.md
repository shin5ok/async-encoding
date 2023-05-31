# Processing each movie asynchronously according to client requests and distribute it

## Prerequiste
- Google Cloud Account
- Google Cloud Project enabled Google Cloud billing
- Terraform v1.4.6 or later  
(See [here](https://developer.hashicorp.com/terraform/downloads) to install the latest one)
- Docker running on your machine

## Procedure to setup the whole system
### 1. Prepare Google Cloud Project and environment variables
Log in to your project,
```
gcloud auth login --update-adc
gcloud config set project <your project id>
```
And then, set some environment variables to be used in the following step.
```
export TF_VAR_domain=<your domain name>
export TF_VAR_region=<GCP region ex:asia-northeast1>
export TF_VAR_zone=<GCP zones splited by comma ex:asia-northeast1-a,asia-northeast1-b>
export TF_VAR_gcs=<your bucket name>
export GOOGLE_CLOUD_PROJECT=<your project id>
```
Note: Domain name must be one you manage.  
If you don't have any domains, You can get your favorite one at *Cloud Domains*.

### 2. Build infrastructure
Preparation and confimation to run totally.
```
cd terraform/
rm -f *tfstate*
terraform plan
```
Just type it to build the whole infrastructure.
```
terraform apply
```

You need to configure DNS A record to assign IP address to your domain.  
There are different ways.  
**After doing that, enabling managed certs may be taking over 10 miniutes.**

Here's the command to confirm if the enablement has been completed.
```
gcloud compute ssl-certificates list
```
Once you see 'ACTIVE' on 'MANAGED_STATUS' line, you can expect to access apis with your domains.  
You might see some error until provisioning has been done completely for a couple of minutes.

### 3. Build applications and deploy them to Cloud Run.  
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
export BASE_URL="https://${TF_VAR_domain}"
make delivering
```

### 4. Prepare to test to work
- Prepare some test movies as MPEG4 format.  
Over 1 mins movie would be nice.  

*Note: You must make sure where you can use the contents.*  
*DO NOT use ones you don't have right to edit.*

- Transfer them to GCS bucket you prepared in advance.   
If you use gcloud cli,
```
gcloud storage cp *.mp4 gs://$TF_VAR_gcs/
```

- List your transfered movies's name into config named movies.txt
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

###  5. Test it
- Make a lot of requests.
Build a command to test.  
If you don't have 'go' command, install the latest one according to [here](https://go.dev/doc/install).
Do test.
```
cd clients/request-clients/
POST_URL="${BASE_URL}/request"
go run . -posturl=$POST_URL -listfile ../../movies.txt -procnum 10 -requestnum 10000
```
This is an example to send 10000 messages as request contains source image and cutting time range randomly, from 10 virtual clients parallelly.


- See Cloud Logging, that will show how processing runs.  
And then, you also see the progress in Firestore and GCS.

- Open the site url that was assigned to **delivering** Cloud Run with your browser.  
Click some links to see movies.  

- Download stored movies parallelly using command line  
```
cd clients/deliver-requests/
LIST_URL="${BASE_URL}/user?jsoned=1"
go run . -listurl=$LIST_URL -movieurl=$BASE_URL -procnum 10
```
This example shows how you can download movies with parallelism 10.  

## 6. Cleanup
See [here](https://cloud.google.com/resource-manager/docs/creating-managing-projects?shutting_down_projects&hl=ja#shutting_down_projects).
