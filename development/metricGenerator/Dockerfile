FROM python:3.9-alpine

WORKDIR /app

COPY . /app

RUN pip install --no-cache-dir --find-links=/app/packages -r requirements.txt

EXPOSE 5001

CMD ["python", "flask_app.py"]