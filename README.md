# Dyndo

Dyndo is a small dynamic DNS client for [Digitalocean](https://www.digitalocean.com/) that uses the offical [API bindings](https://github.com/digitalocean/godo)

## Download
Download release for your architecture


**ARM for Raspiberry pi:**
```
sudo wget https://github.com/mooneyow/dyndo/releases/download/v0.1/dyndo.armv5 -O /usr/local/bin/dyndo
```
**For x86:**
```
sudo wget https://github.com/mooneyow/dyndo/releases/download/v0.1/dyndo.x86_64 -O /usr/local/bin/dyndo
```
## Installation

First create a system user:
```
sudo useradd -r dyndo
```
Then download the service file:
```
sudo wget https://raw.githubusercontent.com/mooneyow/dyndo/master/dyndo.service -O /etc/systemd/system/dyndo.service
```
Add your API key to this file (replacing API_KEY with it):
```
sudo sed -i s/temp_key/API_KEY_HERE/ /etc/systemd/system/dyndo.service
```
Add your domain too (replacing DOMAIN_HERE with it):
```
sudo sed -i s/temp_domain/DOMAIN_HERE/ /etc/systemd/system/dyndo.service
```
Set restrictive permissions on the service file:
```
sudo chmod 600 /etc/systemd/system/dyndo.service
```
Set execute permissions on the binary:
```
sudo chmod +x /usr/local/bin/dyndo
```
Then start and enable the service:
```
sudo systemctl start dyndo && sudo systemctl enable dyndo
```
To check the status:
```
sudo systemctl status dyndo
```

## FAQ

**How often does dyndo check if the address is up to date?**

By default it checks every 5 minutes, this can be adjusted with the `-duration` flag


**Does dyndo support IPv6?**

No, not yet


**Does dyndo support other DNS providers like Cloudflare?**

No, not yet
