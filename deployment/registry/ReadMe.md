
## STEPS TO SETUP PRIVATE DOCKER REGISTRY:

# Registry setup in Linux:

mkdir -p /opt/registry/certs

sudo openssl req -newkey rsa:4096 -nodes -sha256 -keyout /opt/registry/certs/domain.key -x509 -days 365 -out /opt/registry/certs/domain.crt

cd /opt/registry/
sudo mkdir auth

sudo htpasswd -Bc /opt/registry/auth/htpasswd <username>

Eg, username: logesh | pwd: Dev0psU$er

docker-compse up -d

# copy the domain.crt file to host machine for pushing/pulling images.

linux nodes: /etc/docker/certs.d/<Registry_IP>:5000/
windows: C:\ProgramData\DockerDesktop\certs.d\<Registry-IP>_5000\


# Testing

docker login <registry-IP>:5000


-------------------

Openssl cert generation:

cat openssl/san.ext
subjectAltName=IP:54.159.21.131

sudo openssl genpkey -algorithm RSA -out domain.key -aes256
.......+....+...+.....+.+.....+....+.....+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*..+.....+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*.+...+.....+...+....+...+......+.....+............+....+..............+....+...............+...+..+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
..+..................+.+.....+.+...+...+.....+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*.+........+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*.....+...+.....+.+.........+..+...+............+.........+....+..+.+.........+..............+......+...+...+.......+..+....+..+.......+...+.....+..................+......+.+.........+.....+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
Enter PEM pass phrase:
Verifying - Enter PEM pass phrase:
ubuntu@ip-172-31-25-192:~/openssl$ sudo openssl req -new -key domain.key -out domain.csr -subj "/CN=localhost"
Enter pass phrase for domain.key:
ubuntu@ip-172-31-25-192:~/openssl$ ls
domain.csr  domain.key  san.ext
ubuntu@ip-172-31-25-192:~/openssl$ sudo openssl x509 -req -in domain.csr -out domain.crt -signkey domain.key -days 365 -extfile san.ext
Enter pass phrase for domain.key:
Certificate request self-signature ok
subject=CN = localhost
ubuntu@ip-172-31-25-192:~/openssl$ ls
domain.crt  domain.csr  domain.key  san.ext
ubuntu@ip-172-31-25-192:~/openssl$ openssl x509 -in domain.crt -text -noout
Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number:
            4f:d7:a6:34:d2:b2:b7:d9:f1:4d:ce:90:44:f9:cd:54:0f:6f:ad:5b
        Signature Algorithm: sha256WithRSAEncryption
        Issuer: CN = localhost
        Validity
            Not Before: Dec 25 17:25:31 2024 GMT
            Not After : Dec 25 17:25:31 2025 GMT
        Subject: CN = localhost
        Subject Public Key Info:
            Public Key Algorithm: rsaEncryption
                Public-Key: (2048 bit)
                Modulus:
                    00:a3:8c:74:e1:eb:86:af:d1:fb:7c:26:ef:15:f5:
                    03:72:3a:1e:9d:39:ef:18:42:c4:53:52:70:93:60:
                    82:21:a4:82:70:2e:20:75:4c:70:86:d7:13:d3:d1:
                    1a:2c:45:a3:ef:8a:9e:e6:4e:a5:3b:07:e6:17:e6:
                    9f:91:9a:a2:e9:b8:65:e0:eb:4c:dd:66:20:81:e8:
                    05:b5:c3:fd:9d:6f:8a:7e:09:c8:b5:cb:da:f1:7b:
                    99:4c:2f:95:4b:33:4b:2e:fa:73:b1:3b:37:82:36:
                    f6:01:6e:d2:92:a2:05:b4:f5:45:a3:10:95:2f:4e:
                    db:06:a0:4d:b0:90:b0:15:1a:25:53:b0:8a:0b:4b:
                    1e:a5:66:28:e8:09:84:0e:5d:c1:0a:bd:2c:49:9a:
                    76:5a:38:bf:44:a5:ce:19:e5:5b:f0:8e:cd:e2:14:
                    d9:07:c6:17:66:35:5a:ce:07:07:e8:4a:db:c2:5a:
                    c9:80:97:ca:a8:a7:c1:51:9f:f8:db:48:a5:ec:55:
                    9e:1b:f5:61:de:c7:94:66:74:10:55:15:0f:77:08:
                    32:92:1f:79:54:3d:a2:b4:06:d1:35:93:ea:f6:c0:
                    dc:24:3f:87:84:74:f2:cf:36:11:7a:4b:99:1a:95:
                    d3:db:35:70:8a:8a:dc:5a:b4:ea:37:91:b6:1d:28:
                    3f:db
                Exponent: 65537 (0x10001)
        X509v3 extensions:
            X509v3 Subject Alternative Name:
                IP Address:54.159.21.131
            X509v3 Subject Key Identifier:
                04:11:AC:DF:AB:80:C3:31:51:93:9E:BF:4F:31:54:EE:9F:31:20:B6
    Signature Algorithm: sha256WithRSAEncryption
    Signature Value:
        40:06:e4:e0:6b:0f:64:6c:b9:fe:5b:44:64:a8:5c:9b:cf:fe:
        01:e3:04:e3:3a:2a:d7:6c:25:a6:68:6b:60:94:19:2e:63:8d:
        f3:e2:a0:93:db:f0:14:c3:59:06:b0:10:b7:79:88:5f:97:e2:
        6d:32:76:b4:51:7b:54:a7:66:ee:02:bd:1f:17:9e:1f:5a:73:
        11:71:0b:4d:a7:2d:90:5c:59:ec:22:5c:6a:00:54:ca:0f:d4:
        45:35:b4:65:9d:5e:94:21:46:6e:e9:db:8d:7b:bb:09:c5:85:
        27:3f:78:b2:0c:f9:62:d2:81:dd:53:32:c9:6c:86:d5:b5:f9:
        c2:ef:f0:dc:38:ab:6a:20:56:c1:2d:f3:e5:00:d3:9f:95:f5:
        61:df:18:1a:8c:ff:59:2a:22:75:16:44:e6:fc:76:c5:d0:bd:
        88:be:77:c9:94:11:a7:78:33:2a:a7:27:6a:d5:bf:86:3b:f0:
        2d:fe:be:7b:12:06:4d:f0:4b:d8:45:c1:32:7d:2a:97:f7:f2:
        7a:08:51:fb:83:2b:58:23:44:61:8b:18:5d:46:19:eb:d1:70:
        00:b9:55:ef:8b:45:eb:d6:46:06:d3:dc:67:63:22:4a:8b:90:
        05:9d:86:8d:eb:46:f8:64:20:8a:b5:fb:b3:3f:ff:97:92:77:
        e9:54:89:50
ubuntu@ip-172-31-25-192:~/openssl$


sudo chmod 644 /opt/registry/certs/domain.key
ubuntu@ip-172-31-25-192:/opt/registry/certs$ sudo chmod 644 /opt/registry/certs/domain.crt

openssl rsa -in domain.key -out domain.key.unencrypted

openssl rsa -in domain.key.unencrypted -check


mv domain.key.unencrypted domain.key
