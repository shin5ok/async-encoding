from fastapi import FastAPI, Depends, Header, Request
import os
import sys
import uvicorn
import logging
import json
from pydantic import BaseModel, Field, EmailStr
from typing import List, Optional
from fastapi.responses import RedirectResponse, Response
from fastapi.templating import Jinja2Templates
from google.cloud import pubsub_v1


logger = logging.getLogger(__name__)
handler = logging.StreamHandler(sys.stdout)
logger.addHandler(handler)
logger.setLevel(logging.DEBUG)

PORT = os.environ.get("PORT", "8080")
PROJECT_ID = os.environ.get("GOOGLE_CLOUD_PROJECT")
TOPIC = os.environ.get("TOPIC")

class ProcessRequest(BaseModel):
    id: str
    src: str
    start: int
    end: int

app = FastAPI()
publisher = pubsub_v1.PublisherClient()
topic_name = f'projects/{PROJECT_ID}/topics/{TOPIC}'

@app.get("/test")
def _check():
    return {}

@app.post("/request")
def _request(
        request: Request, 
        process_request: ProcessRequest, 
        response: Response,
        user_agent = Header(default=None), 
        host = Header(default=None)
    ):
    print(process_request.dict())

    try:
        data = json.dumps(process_request.dict())
        future = publisher.publish(topic_name, data.encode('utf8'))
        future.result()
    except Exception as e:
        print("Exception:", str(e))
        response.status_code = 500
        return {} 

    return process_request

if __name__ == '__main__':
    port = os.environ.get("PORT", PORT)
    options = {
            'port': int(port),
            'host': '0.0.0.0',
            'workers': 8,
            'reload': True,
        }
    uvicorn.run("main:app", **options)
