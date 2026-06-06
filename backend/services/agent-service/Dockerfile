FROM python:3.11-slim

WORKDIR /app

# System dependencies for OpenVINO + PDF extraction
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    g++ \
    libgl1 \
    libglib2.0-0 \
    libgomp1 \
    ocl-icd-libopencl1 \
    opencl-headers \
    && rm -rf /var/lib/apt/lists/*

# Create and activate virtualenv
RUN python -m venv /app/venv
ENV PATH="/app/venv/bin:$PATH"

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY . .

# Models and vector_db are mounted as volumes — not baked into image
# to keep image size manageable

EXPOSE 8001

CMD ["python", "main.py"]