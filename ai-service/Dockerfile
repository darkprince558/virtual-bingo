FROM python:3.12-slim

WORKDIR /app

# Install dependencies first for better caching
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application source code
COPY . .

# Expose the Flask port
EXPOSE 5000

# Run the Flask app
CMD ["flask", "--app", "app.main:app", "run", "--host=0.0.0.0", "--port=5000"]
