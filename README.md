# Chia Tools

Collection of CLI tools for working with Chia Blockchain

## Installation

Download the correct executable file from the release page and run. If you are on debian/ubuntu, you can install using the apt repo, documented below.

### Apt Repo Installation

#### Set up the repository

1. Update the `apt` package index and install packages to allow apt to use a repository over HTTPS:

```shell
sudo apt-get update

sudo apt-get install ca-certificates curl gnupg
```

2. Add Chia's official GPG Key:

```shell
curl -sL https://repo.chia.net/FD39E6D3.pubkey.asc | sudo gpg --dearmor -o /usr/share/keyrings/chia.gpg
```

3. Use the following command to set up the stable repository.

```shell 
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/chia.gpg] https://repo.chia.net/chia-tools/debian/ stable main" | sudo tee /etc/apt/sources.list.d/chia-tools.list > /dev/null
```

#### Install Chia Tools

1. Update the apt package index and install the latest version of Chia Tools

```shell
sudo apt-get update

sudo apt-get install chia-tools
```
