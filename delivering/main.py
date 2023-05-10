from fastapi import FastAPI, Depends, Header, Request
import os
import sys
import uvicorn
import logging
import json
from pydantic import BaseModel, Field, EmailStr
from fastapi.responses import RedirectResponse, Response
from google.cloud import firestore

COLLECTION = os.environ.get("COLLECTION", "data")
BASE_HOST = os.environ.get("BASE_HOST", "example.com")
PORT = os.environ.get("PORT", "8080")

logger = logging.getLogger(__name__)
handler = logging.StreamHandler(sys.stdout)
logger.addHandler(handler)
logger.setLevel(logging.DEBUG)

app = FastAPI()
db = firestore.Client()

def get_coll():
    docs = db.collection(COLLECTION)
    return docs

class Params(BaseModel):
    src: str
    dst: str
    start: float
    end: float
    user_id: str

@app.get("/user/{user_id}")
def _user_get(user_id: str, request: Request, user_agent = Header(default=None), host = Header(default=None), s = Depends(get_coll)):
    doc = s.document(user_id)
    print(doc.get().to_dict())
    dst = doc.get().get("Dst")
    if not dst:
        return Response(status_code=404)
    url = f"https://{BASE_HOST}/{dst}"
    return RedirectResponse(url, status_code=301)


if __name__ == '__main__':
    port = os.environ.get("PORT", PORT)
    options = {
            'port': int(port),
            'host': '0.0.0.0',
            'workers': 2,
            'reload': True,
        }
    uvicorn.run("main:app", **options)
