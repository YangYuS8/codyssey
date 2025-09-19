# Python AI service Dockerfile
FROM python:3.12-slim
WORKDIR /app
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1
COPY pyproject.toml ./
RUN pip install --upgrade pip && \
    pip install .[dev]
COPY . .
EXPOSE 8000
ENV PY_BACKEND_PORT=8000
CMD ["python", "main.py"]
