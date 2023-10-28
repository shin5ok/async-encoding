from fastapi import FastAPI, Depends, Header, Request
import os
import sys
import uvicorn
import logging
import json
from pydantic import BaseModel, Field, EmailStr
from typing import List
from fastapi.responses import RedirectResponse, Response
from fastapi.templating import Jinja2Templates
from google.cloud import pubsub_v1


logger = logging.getLogger(__name__)
handler = logging.StreamHandler(sys.stdout)
logger.addHandler(handler)
logger.setLevel(logging.DEBUG)

PROJECT_ID = os.environ.get("PROJECT_ID")
PORT = os.environ.get("PORT", "8080")
TOPIC = os.environ.get("TOPIC")

class ProcessRequest(BaseModel):
    user_id: str
    src: str
    start: int
    end: int

app = FastAPI()
publisher = pubsub_v1.PublisherClient()
topic_name = f'projects/{PROJECT_ID}/topics/{TOPIC}'

@app.get("/test")
def _check() -> dict:
    return {}

@app.post("/request")
def _request(
        request: Request, 
        process_request: ProcessRequest, 
        response: Response,
        user_agent = Header(default=None), 
        host = Header(default=None)
    ) -> ProcessRequest | dict:
    print(process_request.model_dump())

    try:
        data = json.dumps(process_request.model_dump())
        future = publisher.publish(topic_name, data.encode('utf8'))
        future.result()
    except Exception as e:
        print("Exception:", str(e))
        response.status_code = 500
        return {} 

    return process_request

if __name__ == '__main__':
    options = {
            'port': int(PORT),
            'host': '0.0.0.0',
            'workers': 8,
            'reload': True,
        }
    uvicorn.run("main:app", **options)
