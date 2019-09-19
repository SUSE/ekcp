#!/bin/bash
set -ex

# Usage: DEPLOY_MASTER=true SSH_KEY=~/.ssh/id_rsa API_SERVER=user@some.host API_SLAVE="user@host1 user@host2" ./deploy_opensuse.sh
# Do not use it!!!! :)

MASTER_USER="$(cut -d'@' -f1 <<<"$API_SERVER")"
MASTER_IP="$(cut -d'@' -f2 <<<"$API_SERVER")"

function install() {
    local host=$1
    ssh -4 -t -i $SSH_KEY $host 'sudo zypper in -y zsh docker docker-compose git' || true
    ssh -4 -t -i $SSH_KEY $host 'sudo gpasswd -a $USER docker' || true
    ssh -4 -t -i $SSH_KEY $host '[ ! -d ekcp ] && mkdir ekcp' || true
    ssh -4 -t -i $SSH_KEY $host 'sudo systemctl start docker && sudo systemctl enable docker'
    ssh -4 -t -i $SSH_KEY $host 'sudo systemctl stop firewalld && sudo systemctl disable firerwalld' || true
    scp -i $SSH_KEY docker-compose.yaml $host:./ekcp
    ssh -4 -i $SSH_KEY $host 'pushd $HOME/ekcp && docker-compose -f docker-compose.yaml down'  || true
    ssh -4 -i $SSH_KEY $host 'pushd $HOME/ekcp && docker-compose -f docker-compose.yaml up -d'
}

if [ "$DEPLOY_MASTER" == true ]; then
 
cat <<EOF > docker-compose.yaml
version: '3'
services:
  # Master (API) instance
  ekcp-master:
    environment:
      - HOST=0.0.0.0
      - PORT=8030
      - ROUTE_REGISTER=false
      - DOMAIN=nip.io
      - KUBEHOST=${MASTER_IP}  # Tweak this to your lan ip      
      - FEDERATION=true
    image: quay.io/ekcp/ekcp
    network_mode: "host"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    deploy:
      restart_policy:
        condition: always
EOF
    install $API_SERVER
fi

for i in $API_SLAVE; 
do 
    USER="$(cut -d'@' -f1 <<<"$i")"
    IP="$(cut -d'@' -f2 <<<"$i")"

cat <<EOF > docker-compose.yaml
version: '3'
services:
  ekcp-slave:
    environment:
      - HOST=0.0.0.0
      - PORT=8030
      - ROUTE_REGISTER=true
      - DOMAIN=nip.io
      - KUBEHOST=${IP}  # Tweak this to your lan ip      
      - FEDERATION_MASTER=http://${MASTER_IP}:8030
    image: quay.io/ekcp/ekcp
    #build: .
    network_mode: "host"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    deploy:
      restart_policy:
        condition: always
  nats:
    image: nats:latest
    ports:
      - "4222:4222"
      - "6222:6222"
      - "8222:8222"
    deploy:
      restart_policy:
        condition: always
  gorouter:
   # Replace the pinned cert for security.
    image: quay.io/ekcp/gorouter
    ports:
      - "8081:8081"
      - "8082:8082"
      - "8083:8083"
    links:
      - nats
    deploy:
      restart_policy:
        condition: always
EOF

    install $i
done
