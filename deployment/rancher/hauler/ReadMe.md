

## OS: RHEL 9:

## Rancher Manager Setup(Works only in UBUNTU):

sudo apt update -y && sudo apt upgrade -y

https://docs.docker.com/engine/install/ubuntu/

sudo groupadd docker

sudo usermod -aG docker $USER

docker ps

sudo reboot

sudo systemctl start docker

sudo systemctl enable docker

docker ps

sudo docker run --privileged -d --restart=unless-stopped -p 80:80 -p 443:443 rancher/rancher


## Master Node Registeration Steps:

sudo dnf update -y

curl https://releases.rancher.com/install-docker/20.10.sh | sh

sudo systemctl disable nm-cloud-setup.service nm-cloud-setup.timer

sudo groupadd docker

sudo usermod -aG docker $USER

docker ps

sudo reboot

sudo systemctl start docker

sudo systemctl enable docker

docker ps

curl --insecure -fL https://184.72.141.58/system-agent-install.sh | sudo  sh -s - --server https://184.72.141.58 --label 'cattle.io/os=linux' --token kd2gq7568gtqbn8gql8jtrf4cggktn4f7tt6nn4lmrhrpsltwblhzn --ca-checksum a241d782705ece86da2ff95e55e7e376707e59e6e486af3733a51ed695b072a5 --etcd --controlplane

systemctl status rancher-system-agent.service


## Worker Node Registeration Steps:

sudo dnf update -y

curl https://releases.rancher.com/install-docker/20.10.sh | sh

sudo systemctl disable nm-cloud-setup.service nm-cloud-setup.timer

sudo groupadd docker

sudo usermod -aG docker $USER

docker ps

sudo reboot

sudo systemctl start docker

sudo systemctl enable docker

docker ps

curl --insecure -fL https://184.72.141.58/system-agent-install.sh | sudo  sh -s - --server https://184.72.141.58 --label 'cattle.io/os=linux' --token kd2gq7568gtqbn8gql8jtrf4cggktn4f7tt6nn4lmrhrpsltwblhzn --ca-checksum a241d782705ece86da2ff95e55e7e376707e59e6e486af3733a51ed695b072a5 --worker

systemctl status rancher-system-agent.service


# Kubectl cmds:


