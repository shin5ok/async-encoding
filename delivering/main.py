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
import jinja2

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

class Item(BaseModel):
    dst: str
    user_id: str
    key: str

class ItemList(BaseModel):
    lists: None | List[Item]

@app.get("/user/{user_id}")
def _user_get(user_id: str, request: Request, user_agent = Header(default=None), host = Header(default=None), s = Depends(get_coll)):
    doc = s.document(user_id)
    print(doc.get().to_dict())
    dst = doc.get().get("Dst")
    if not dst:
        return Response(status_code=404)
    url = f"https://{BASE_HOST}/{dst}"
    return RedirectResponse(url, status_code=301)

@app.get("/user")
def _user(request: Request, user_agent = Header(default=None), host = Header(default=None), s = Depends(get_coll)):
    r = s.limit(100)
    lists = []
    for x in r.stream():
        d = x.to_dict()
        lists.append(Item(dst=d.get('Dst'), user_id=d.get('UserID'), key=x.id))
    return Jinja2Templates(directory="templates").TemplateResponse("item_list.html", dict(lists=lists, request=request))

if __name__ == '__main__':
    port = os.environ.get("PORT", PORT)
    options = {
            'port': int(port),
            'host': '0.0.0.0',
            'workers': 2,
            'reload': True,
        }
    uvicorn.run("main:app", **options)
