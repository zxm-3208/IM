#!/bin/bash
reso_addr='registry.cn-hangzhou.aliyuncs.com/my-im/social-rpc-dev'
tag='latest'

pod_ip="139.9.214.194"

container_name="im-social-rpc-test"

docker stop ${container_name}

docker rm ${container_name}

docker rmi ${reso_addr}:${tag}

docker pull ${reso_addr}:${tag}


# 如果需要指定配置文件的
# docker run -p 10001:8080 --network easy-chat -v /easy-chat/config/user-rpc:/user/conf/ --name=${container_name} -d ${reso_addr}:${tag}
docker run -p 10001:10001 -e POD_IP=${pod_ip} --name=${container_name} -d ${reso_addr}:${tag}
