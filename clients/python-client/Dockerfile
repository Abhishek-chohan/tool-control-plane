# Use the official Ubuntu base image
FROM ubuntu:latest

# Set the working directory
WORKDIR /app

# Install Python and pip
RUN apt-get update && apt-get install -y \
    python3 \
    python3-pip \
    python3-venv \
    && apt-get clean

# Create a virtual environment
RUN python3 -m venv /app/venv

# Activate the virtual environment and install the wheel file
COPY dist/toolplane-0.1.0-py3-none-any.whl /app/toolplane-0.1.0-py3-none-any.whl
RUN /app/venv/bin/pip install --no-cache-dir /app/toolplane-0.1.0-py3-none-any.whl

# Copy the requirements.txt file into the container
COPY requirements.txt /app/requirements.txt
RUN /app/venv/bin/pip install --no-cache-dir -r /app/requirements.txt

# Copy the example.py script into container
COPY example.py /app/example.py

# Set up the folder to be mounted as root
VOLUME /root

# Ensure the container runs as root
USER root

# Run the example.py script using the virtual environment
CMD ["sh", "-c", "cd /root && /app/venv/bin/python /app/example.py \"$SESSION_ID\" \"$USER_ID\" \"$API_KEY\""]