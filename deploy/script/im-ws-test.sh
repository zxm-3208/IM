#!/bin/bash
reso_addr='registry.cn-hangzhou.aliyuncs.com/my-im/im-ws-dev'
tag='latest'

container_name="im-im-ws-test"

docker stop ${container_name}

docker rm ${container_name}

docker rmi ${reso_addr}:${tag}

docker pull ${reso_addr}:${tag}


# 如果需要指定配置文件的
# docker run -p 10001:8080 --network easy-chat -v /easy-chat/config/user-rpc:/user/conf/ --name=${container_name} -d ${reso_addr}:${tag}
docker run -p 10090:10090 --name=${container_name} -d ${reso_addr}:${tag}
