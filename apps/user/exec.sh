# gRPC 文件生成
goctl rpc protoc ./apps/user/rpc/user.proto --go_out=./apps/user/rpc/ --go-grpc_out=./apps/user/rpc/ --zrpc_out=./apps/user/rpc/

# mysql model文件生成
goctl model mysql ddl -src="./deploy/sql/user.sql" -dir="./apps/user/models/" -c

# 构建user api服务
goctl api go -api apps/user/api/user.api -dir apps/user/api -style gozero