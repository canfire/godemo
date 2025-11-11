FROM registry.yygu.cn/base/python:3.10
RUN pip install fastapi uvicorn pandas
COPY sandbox.py /app/sandbox.py
WORKDIR /app
CMD ["uvicorn", "sandbox:app", "--host", "0.0.0.0", "--port", "8080"]
