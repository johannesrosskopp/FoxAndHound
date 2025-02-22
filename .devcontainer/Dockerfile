FROM mcr.microsoft.com/devcontainers/go:1-1.23-bookworm

# Install Azure CLI
RUN curl -sL https://aka.ms/InstallAzureCLIDeb | bash

# RUN apt-get update && sudo apt-get install -y gnupg software-properties-common

# Installing terraform
RUN wget -O- https://apt.releases.hashicorp.com/gpg | \
    gpg --dearmor | tee /usr/share/keyrings/hashicorp-archive-keyring.gpg > /dev/null \
    && echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] \
        https://apt.releases.hashicorp.com $(lsb_release -cs) main" | \
        sudo tee /etc/apt/sources.list.d/hashicorp.list

RUN apt update && apt install terraform