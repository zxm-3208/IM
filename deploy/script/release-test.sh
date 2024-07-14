need_start_server_shell=(

  user-rpc-test.sh
  social-rpc-test.sh
  im-rpc-test.sh

  user-api-test.sh
  social-api-test.sh
  im-api-test.sh
  im-ws-test.sh
  task-mq-test.sh
)

for i in ${need_start_server_shell[*]}; do
  chmod +x $i
  ./$i
done

docker ps

docker exec -it etcd etcdctl get --prefix ""