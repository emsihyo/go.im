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

docker rm -rf go.im.server
docker pull registry.cn-shenzhen.aliyuncs.com/emsihyo/go.im.server
docker rmi $(docker images -f "dangling=true" -q)
docker run -d --name go.im.server -p 10001:10001 -p 10002:10002 registry.cn-shenzhen.aliyuncs.com/emsihyo/go.im.server

