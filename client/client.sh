#!/bin/sh
sudo yum install -y yum-utils device-mapper-persistent-data lvm2
sudo yum-config-manager --add-repo http://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
sudo yum makecache fast
sudo yum -y install docker-ce
sudo mkdir -p /etc/docker
sudo tee /etc/docker/daemon.json <<-'EOF'
{
  "registry-mirrors": ["https://cfcbfldf.mirror.aliyuncs.com"]
}
EOF
sudo systemctl daemon-reload
sudo systemctl restart docker
docker rm -f go.im.client
docker pull registry.cn-shenzhen.aliyuncs.com/emsihyo/go.im.client
docker rmi $(docker images -f "dangling=true" -q)
docker run -d --name go.im.client -p 10000:10000 registry.cn-shenzhen.aliyuncs.com/emsihyo/go.im.client  --port=10000 --host=172.18.82.132:10001 --users=10000 --topics=1000 --per=1 --duration=60
