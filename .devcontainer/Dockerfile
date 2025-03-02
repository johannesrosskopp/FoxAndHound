FROM mcr.microsoft.com/devcontainers/base:bullseye

# Install dependencies
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    gnupg \
    software-properties-common

# Install Go 1.23.5
RUN wget https://go.dev/dl/go1.23.5.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go1.23.5.linux-amd64.tar.gz \
    && rm go1.23.5.linux-amd64.tar.gz

ENV PATH "$PATH:/usr/local/go/bin"

# Install Azure CLI
RUN curl -sL https://aka.ms/InstallAzureCLIDeb | bash

# RUN apt-get update && sudo apt-get install -y gnupg software-properties-common

# Installing terraform
# RUN wget -O- https://apt.releases.hashicorp.com/gpg | \
#     gpg --dearmor | tee /usr/share/keyrings/hashicorp-archive-keyring.gpg > /dev/null \
#     && echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] \
#         https://apt.releases.hashicorp.com $(lsb_release -cs) main" | \
#         sudo tee /etc/apt/sources.list.d/hashicorp.list

# RUN apt update && apt install terraform

# Installing pulumi

RUN curl -fsSL https://get.pulumi.com | sh 
RUN mkdir /usr/bin/.pulumi && mv /root/.pulumi/* /usr/bin/.pulumi/ 
ENV PATH "$PATH:/usr/bin/.pulumi/bin/"