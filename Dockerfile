FROM python:3.11-slim

# Set working directory
WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir --trusted-host pypi.org --trusted-host pypi.python.org --trusted-host files.pythonhosted.org -r requirements.txt

# Copy application code
COPY *.py ./
COPY config.json ./

# Create output directory
RUN mkdir -p /app/output

# Set environment variable to prevent Python from writing .pyc files
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

# Default command shows help
ENTRYPOINT ["python", "main.py"]
CMD ["--help"]
