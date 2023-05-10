FROM python:3.10-slim

COPY . .
RUN pip install poetry; poetry config virtualenvs.in-project true
RUN poetry install

CMD ["poetry", "run", "python", "main.py"]
