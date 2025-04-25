FROM python:3.10-slim

WORKDIR /app

# Copy requirements file
COPY requirements.txt .

# Install dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy the script
COPY generate_court_model.py .

# Create output directory
RUN mkdir -p /app/output

# Set the entrypoint
ENTRYPOINT ["python", "generate_court_model.py"]

# Default command (can be overridden)
CMD [] 
