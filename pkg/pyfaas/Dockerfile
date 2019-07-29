FROM python:3.7-alpine as base
FROM base as builder
RUN mkdir /install
RUN pip install pipreqs
WORKDIR /install
COPY main.py .
RUN pipreqs --force --save /requirements.txt .
RUN pip install --install-option="--prefix=/install" -r /requirements.txt

FROM base
COPY --from=builder /install /usr/local
COPY . /app
WORKDIR /app
EXPOSE 5000
ENTRYPOINT ["python", "/app/main.py"]
