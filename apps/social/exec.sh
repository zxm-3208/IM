# gRPC 文件生成
goctl rpc protoc ./apps/social/rpc/social.proto --go_out=./apps/social/rpc --go-grpc_out=./apps/social/rpc --zrpc_out=./apps/social/rpc

# mysql model文件生成
goctl model mysql ddl -src="./deploy/sql/social.sql" -dir="./apps/social/socialmodels/" -c

# api 文件生成
goctl api go -api apps/social/api/social.api -dir apps/social/api -style gozero
